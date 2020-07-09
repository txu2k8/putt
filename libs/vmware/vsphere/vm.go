package vsphere

// VM: object.VirtualMachine

import (
	"context"
	"fmt"
	"putt/libs/convert"
	"putt/libs/retry"
	"putt/libs/retry/strategy"
	"sync"
	"time"

	"github.com/chenhg5/collection"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// Worker ...
type Worker struct {
	wg          sync.WaitGroup
	done        chan struct{}
	maxParallel int
}

// OptVM . vm object opt map
type OptVM struct {
	Opt string                 // poweron/poweroff/shutdown/reset/reboot/suspend
	VM  *object.VirtualMachine // vm object
}

// =============== Get VM Properties ===============

// VMGuestInfo .
func VMGuestInfo(vm *object.VirtualMachine) *types.GuestInfo {
	var o mo.VirtualMachine
	ctx := context.Background()
	err := vm.Properties(ctx, vm.Reference(), []string{"guest"}, &o)
	if err != nil {
		logger.Fatal(err)
	}
	if o.Guest == nil {
		logger.Fatal("Guest=nil")
	}
	return o.Guest
}

// VMConfigInfo .
func VMConfigInfo(vm *object.VirtualMachine) *types.VirtualMachineConfigInfo {
	var o mo.VirtualMachine
	ctx := context.Background()
	err := vm.Properties(ctx, vm.Reference(), []string{"config"}, &o)
	if err != nil {
		logger.Fatal(err)
	}
	if o.Config == nil {
		logger.Fatal("Config=nil")
	}
	return o.Config
}

// VMNetworkInfo .
func VMNetworkInfo(vm *object.VirtualMachine) []types.ManagedObjectReference {
	var o mo.VirtualMachine
	ctx := context.Background()
	err := vm.Properties(ctx, vm.Reference(), []string{"network"}, &o)
	if err != nil {
		logger.Fatal(err)
	}
	if o.Network == nil {
		logger.Fatal("Config=nil")
	}
	return o.Network
}

// GetVMUUID .
func GetVMUUID(vm *object.VirtualMachine) string {
	ctx := context.Background()
	return vm.UUID(ctx)
}

// GetVMName .
func GetVMName(vm *object.VirtualMachine) string {
	return VMConfigInfo(vm).Name

}

// GetVMIP .
func GetVMIP(vm *object.VirtualMachine) string {
	return VMGuestInfo(vm).IpAddress
}

// GetVMHostName .
func GetVMHostName(vm *object.VirtualMachine) string {
	return VMGuestInfo(vm).HostName
}

// GetVMNetworkName .TODO
func GetVMNetworkName(vm *object.VirtualMachine) string {
	return VMNetworkInfo(vm)[0].Reference().String()
}

// =============== Power VM ===============

// IsVMPowerStateExpected .
func IsVMPowerStateExpected(vm *object.VirtualMachine, state types.VirtualMachinePowerState) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	curState, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("%s powerState(runtime/expected):%s/%s", vmName, curState, state)
	if curState == state {
		logger.Infof(msg)
		return nil
	}
	return fmt.Errorf(msg)
}

// WaitForVMPowerState .
func WaitForVMPowerState(vm *object.VirtualMachine, state types.VirtualMachinePowerState, tries int) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	logger.Infof("Wait For VM %s: %s ...", vmName, state)
	action := func(attempt uint) error {
		return vm.WaitForPowerState(ctx, state)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
	)
	curState, _ := vm.PowerState(ctx)
	logger.Infof("%s powerState(runtime/expected):%s/%s", vmName, curState, state)
	return err
}

// PowerOffVM .
func PowerOffVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStatePoweredOff {
		logger.Infof("%s already poweredOff", vmName)
		return nil
	}

	logger.Infof("PowerOff %s ...", vmName)
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	err = WaitForVMPowerState(vm, types.VirtualMachinePowerStatePoweredOff, 30)
	if err != nil {
		return err
	}
	return nil
}

// PowerOnVM .
func PowerOnVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStatePoweredOn {
		logger.Infof("%s already poweredOn", vmName)
		return nil
	}

	logger.Infof("PowerOn %s ...", vmName)
	task, err := vm.PowerOn(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	err = WaitForVMPowerState(vm, types.VirtualMachinePowerStatePoweredOn, 30)
	if err != nil {
		return err
	}
	return nil
}

// ShutdownVM .
func ShutdownVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStatePoweredOff {
		logger.Infof("%s already poweredOff", vmName)
	}
	logger.Infof("ShutdownGuest %s ...", vmName)
	err = vm.ShutdownGuest(ctx)
	if err != nil {
		return err
	}
	err = WaitForVMPowerState(vm, types.VirtualMachinePowerStatePoweredOff, 30)
	if err != nil {
		return err
	}
	return nil
}

// ResetVM .
func ResetVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStatePoweredOff ||
		state == types.VirtualMachinePowerStateSuspended {
		return fmt.Errorf("powerState(%s) VM can not be reboot", state)
	}

	logger.Infof("Reset %s ...", vmName)
	task, err := vm.Reset(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	err = WaitForVMPowerState(vm, types.VirtualMachinePowerStatePoweredOn, 30)
	if err != nil {
		return err
	}
	return nil
}

// RebootVM .
func RebootVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStatePoweredOff ||
		state == types.VirtualMachinePowerStateSuspended {
		return fmt.Errorf("powerState(%s) VM can not be reboot", state)
	}
	logger.Infof("RebootGuest %s ...", vmName)
	err = vm.RebootGuest(ctx)
	if err != nil {
		return err
	}
	err = WaitForVMPowerState(vm, types.VirtualMachinePowerStatePoweredOn, 30)
	if err != nil {
		return err
	}
	return nil
}

// SuspendVM .
func SuspendVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStateSuspended {
		logger.Infof("%s already suspended", vmName)
		return nil
	} else if state == types.VirtualMachinePowerStatePoweredOff {
		return fmt.Errorf("powerState(%s) VM can not be suspend", state)
	}

	logger.Infof("Suspend %s ...", vmName)
	task, err := vm.Suspend(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	err = WaitForVMPowerState(vm, types.VirtualMachinePowerStateSuspended, 30)
	if err != nil {
		return err
	}
	return nil
}

// DestroyVM .
func DestroyVM(vm *object.VirtualMachine) error {
	ctx := context.Background()
	vmName := GetVMName(vm)
	state, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if state == types.VirtualMachinePowerStatePoweredOn ||
		state == types.VirtualMachinePowerStateSuspended {
		return fmt.Errorf("powerState(%s) VM can not be destroy", state)
	}

	logger.Infof("Destroy %s ...", vmName)
	task, err := vm.Destroy(ctx)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

// PowerOptVM .
func PowerOptVM(vm *object.VirtualMachine, opt string) error {
	var err error
	optState := map[string]types.VirtualMachinePowerState{
		"poweroff": types.VirtualMachinePowerStatePoweredOff,
		"shutdown": types.VirtualMachinePowerStatePoweredOff,
		"poweron":  types.VirtualMachinePowerStatePoweredOn,
		"suspend":  types.VirtualMachinePowerStateSuspended,
		"reset":    types.VirtualMachinePowerStatePoweredOn,
		"reboot":   types.VirtualMachinePowerStatePoweredOn,
	}
	ctx := context.Background()
	vmName := GetVMName(vm)
	curState, err := vm.PowerState(ctx)
	if err != nil {
		return err
	}
	if !collection.Collect([]string{"reset", "reboot"}).Contains(opt) {
		if curState == optState[opt] {
			logger.Infof("%s powerState already %s", vmName, curState)
			return nil
		}
	}

	var task *object.Task
	switch opt {
	case "poweroff":
		task, err = vm.PowerOff(ctx)
	case "shutdown":
		err = vm.ShutdownGuest(ctx)
	case "poweron":
		task, err = vm.PowerOn(ctx)
	case "suspend":
		if curState == types.VirtualMachinePowerStatePoweredOff {
			return fmt.Errorf("powerState(%s) VM can not be suspend", curState)
		}
		task, err = vm.Suspend(ctx)
	case "reset":
		if curState == types.VirtualMachinePowerStatePoweredOff ||
			curState == types.VirtualMachinePowerStateSuspended {
			return fmt.Errorf("powerState(%s) VM can not be reset", curState)
		}
		task, err = vm.Reset(ctx)
	case "reboot":
		if curState == types.VirtualMachinePowerStatePoweredOff ||
			curState == types.VirtualMachinePowerStateSuspended {
			return fmt.Errorf("powerState(%s) VM can not be reboot", curState)
		}
		err = vm.RebootGuest(ctx)
	default:
		logger.Fatalf("Not supported power opt: %s", opt)

	}

	logger.Infof("%s %s ...", convert.StrFirstToUpper(opt), vmName)
	if err != nil {
		return err
	}
	if task != nil {
		err = task.Wait(ctx)
		if err != nil {
			return err
		}
	}

	err = WaitForVMPowerState(vm, optState[opt], 30)
	if err != nil {
		return err
	}
	return nil
}

// MultiPowerOptVM .
func MultiPowerOptVM(optVms []OptVM) error {
	var err error
	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)
	for _, target := range optVms {
		opt := target.Opt
		vm := target.VM
		time.Sleep(2 * time.Second)
		select {
		case ch <- struct{}{}:
			w.wg.Add(1)
			go func() {
				err = PowerOptVM(vm, opt)
				if err != nil {
					w.wg.Done()
					w.done <- struct{}{}
				}
				<-ch
				w.wg.Done()
			}()
		case <-w.done:
			break
		}
	}
	w.wg.Wait()
	return err
}

// =============== VM Snapshot/Clone ===============

// CreateSnapShot Create SnapShot for VM
func CreateSnapShot(vm *object.VirtualMachine, name, desc string) error {
	ctx := context.Background()
	task, err := vm.CreateSnapshot(ctx, name, desc, false, false)
	if err != nil {
		return err
	}
	if err = task.Wait(ctx); err != nil {
		return err
	}
	return nil
}

// RecoverToSnapshot Recover VM To exist Snapshot
func RecoverToSnapshot(vm *object.VirtualMachine, name string, suppressPowerOn bool) error {
	ctx := context.Background()
	task, err := vm.RevertToSnapshot(ctx, name, suppressPowerOn)
	if err != nil {
		return err
	}
	if err = task.Wait(ctx); err != nil {
		return err
	}
	return nil
}

// =============== Config VM ===============

// ResizeVMCPU .
func ResizeVMCPU(vm *object.VirtualMachine, cpuNum int) error {
	ctx := context.Background()
	config := types.VirtualMachineConfigSpec{
		NumCPUs: int32(cpuNum),
	}
	task, err := vm.Reconfigure(ctx, config)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	vmName := GetVMName(vm)
	logger.Infof("Resize VM %s cpu num: %d", vmName, cpuNum)
	return nil
}

// ResizeVMMem ...
func ResizeVMMem(vm *object.VirtualMachine, memMB int) error {
	ctx := context.Background()
	config := types.VirtualMachineConfigSpec{
		MemoryMB: int64(memMB),
	}
	task, err := vm.Reconfigure(ctx, config)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	vmName := GetVMName(vm)
	logger.Infof("Resize VM %s mem(MB):%d", vmName, memMB)
	return nil
}

// ReserveVMCPU ...
func ReserveVMCPU(vm *object.VirtualMachine, cpuMHz int) error {
	ctx := context.Background()
	config := types.VirtualMachineConfigSpec{
		CpuAllocation: &types.ResourceAllocationInfo{
			Reservation: int64Ptr(int64(cpuMHz)),
		},
	}
	task, err := vm.Reconfigure(ctx, config)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	vmName := GetVMName(vm)
	logger.Infof("Reservation VM %s cpu(MHz):%d", vmName, cpuMHz)
	return nil
}

// ReserveVMMem ...
func ReserveVMMem(vm *object.VirtualMachine, memMB int) error {
	ctx := context.Background()
	config := types.VirtualMachineConfigSpec{
		MemoryAllocation: &types.ResourceAllocationInfo{
			Reservation: int64Ptr(int64(memMB)),
		},
	}
	task, err := vm.Reconfigure(ctx, config)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	vmName := GetVMName(vm)
	logger.Infof("Reservation VM %s mem(MB):%d", vmName, memMB)
	return nil
}

// CreateVMDisk . TODO
func CreateVMDisk(vm *object.VirtualMachine) error {
	ctx := context.Background()
	spec := types.VirtualMachineConfigSpec{}
	config := &types.VirtualDeviceConfigSpec{
		Operation:     types.VirtualDeviceConfigSpecOperationAdd,
		FileOperation: types.VirtualDeviceConfigSpecFileOperationCreate,
		Device: &types.VirtualDisk{
			VirtualDevice: types.VirtualDevice{
				Key:           2002,
				ControllerKey: 1000,
				UnitNumber:    int32Ptr(3), // zero default value
				Backing: &types.VirtualDiskFlatVer2BackingInfo{
					DiskMode:        string(types.VirtualDiskModePersistent),
					ThinProvisioned: boolPtr(false),
					VirtualDeviceFileBackingInfo: types.VirtualDeviceFileBackingInfo{
						FileName: "[datastore1]",
					},
				},
			},
			CapacityInKB: 30 * 1024 * 1024,
		},
	}
	spec.DeviceChange = append(spec.DeviceChange, config)
	task, err := vm.Reconfigure(ctx, spec)
	if err != nil {
		return err
	}
	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}
