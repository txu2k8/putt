package vsphere

import (
	"context"
	"fmt"
	"net/url"
	"putt/libs/retry"
	"putt/libs/retry/strategy"
	"putt/libs/utils"
	"time"

	"github.com/op/go-logging"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

var logger = logging.MustGetLogger("test")

// VspConfig vsphere api config
type VspConfig struct {
	Host     string
	Username string
	Password string
	Port     int
}

// VspClient govmomi.Client & rest.Client
type VspClient struct {
	Govmomi *govmomi.Client
	// Rest    *rest.Client
}

// ByName .
type ByName []mo.VirtualMachine

func (n ByName) Len() int           { return len(n) }
func (n ByName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n ByName) Less(i, j int) bool { return n[i].Name < n[j].Name }

// Override username and/or password as required
func processOverride(u *url.URL, user, pwd string) {
	// Override username if provided
	if user != "" {
		var password string
		var ok bool
		if u.User != nil {
			password, ok = u.User.Password()
		}
		if ok {
			u.User = url.UserPassword(user, password)
		} else {
			u.User = url.User(user)
		}
	}

	// Override password if provided
	if pwd != "" {
		var username string
		if u.User != nil {
			username = u.User.Username()
		}
		u.User = url.UserPassword(username, pwd)
	}
}

// NewClient create a VspClient
func NewClient(vc *VspConfig) (*VspClient, error) {
	var cli VspClient
	logger.Infof("Connect vsphere:%+v", *vc)
	ctx := context.Background()
	url := vc.Host
	// Parse URL from string
	u, err := soap.ParseURL(url)
	if err != nil {
		return nil, err
	}

	// Override username and/or password as required
	processOverride(u, vc.Username, vc.Password)

	// Connect and log in to ESX or vCenter
	gc, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return &cli, fmt.Errorf("Connecting to govmomi api failed: %w", err)
	}
	cli.Govmomi = gc

	// cli.Rest = rest.NewClient(cli.Govmomi.Client)
	// err = cli.Rest.Login(ctx, u.User)
	// if err != nil {
	// 	return &cli, fmt.Errorf("log in to rest api failed: %w", err)
	// }

	return &cli, nil
}

// NewClientWithRetry create a govmomi.Client
func NewClientWithRetry(vc *VspConfig) (cli *VspClient, err error) {
	interval := time.Duration(15)
	timeout := time.NewTimer(30 * time.Minute)
loop:
	for {
		cli, err = NewClient(vc)
		if err == nil && cli != nil {
			logger.Debug("Connect vsphere API success")
			break loop
		}
		logger.Warningf("New vsphere client failed, %v", err)

		// retry or timeout
		select {
		case <-time.After(interval * time.Second):
			logger.Infof("retry connect vsphere client after %d second", interval)
		case <-timeout.C:
			err = fmt.Errorf("New vsphere client failed after retry many times, cause by %v", err)
			break loop
		}
	}
	return
}

// Run calls f with Client create from the -url flag if provided,
// otherwise runs the example against vcsim.
func Run(vc *VspConfig, f func(context.Context, *vim25.Client) error) {
	var err error
	if vc.Host == "" {
		err = simulator.VPX().Run(f)
	} else {
		ctx := context.Background()
		var c *VspClient
		c, err = NewClientWithRetry(vc)
		if err == nil {
			err = f(ctx, c.Govmomi.Client)
		}
	}
	if err != nil {
		logger.Fatal(err)
	}
}

// Logout .
func (c *VspClient) Logout(ctx context.Context) error {
	err := c.Govmomi.Logout(ctx)
	if err != nil {
		return fmt.Errorf("govmomi api logout failed: %w", err)
	}
	return nil
}

// =============== Client: Get VM-Object ===============

// GetVMByUUID .
func (c *VspClient) GetVMByUUID(uuid string) (*object.VirtualMachine, error) {
	ctx := context.Background()
	searchIndex := object.NewSearchIndex(c.Govmomi.Client)
	reference, err := searchIndex.FindByUuid(ctx, nil, uuid, true, nil)
	if reference == nil {
		return nil, err
	}
	vm := object.NewVirtualMachine(c.Govmomi.Client, reference.Reference())
	return vm, nil
}

// GetVMByIP .
func (c *VspClient) GetVMByIP(vmIP string) (*object.VirtualMachine, error) {
	ctx := context.Background()
	searchIndex := object.NewSearchIndex(c.Govmomi.Client)
	reference, err := searchIndex.FindByIp(ctx, nil, vmIP, true)
	if reference == nil {
		return nil, err
	}
	vm := object.NewVirtualMachine(c.Govmomi.Client, reference.Reference())
	return vm, nil
}

// GetVMByDNSName .
func (c *VspClient) GetVMByDNSName(dnsName string) (*object.VirtualMachine, error) {
	ctx := context.Background()
	searchIndex := object.NewSearchIndex(c.Govmomi.Client)
	reference, err := searchIndex.FindByDnsName(ctx, nil, dnsName, true)
	if reference == nil {
		return nil, err
	}
	vm := object.NewVirtualMachine(c.Govmomi.Client, reference.Reference())
	return vm, nil
}

// GetVMByName .
func (c *VspClient) GetVMByName(vmName string) (*object.VirtualMachine, error) {
	ctx := context.Background()
	searchIndex := object.NewSearchIndex(c.Govmomi.Client)
	childEntity, _ := object.NewRootFolder(c.Govmomi.Client).Children(ctx)
	for _, child := range childEntity {
		reference, _ := searchIndex.FindChild(ctx, child, vmName)
		if reference == nil {
			continue
		}
		vm := object.NewVirtualMachine(c.Govmomi.Client, reference.Reference())
		return vm, nil
	}
	return nil, fmt.Errorf("Got none VM with vm-name:%s", vmName)
}

// GetVMDetails .
func (c *VspClient) GetVMDetails(vmName string) (*mo.VirtualMachine, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create view of VirtualMachine objects
	m := view.NewManager(c.Govmomi.Client)

	v, err := m.CreateContainerView(ctx, c.Govmomi.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all machines
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.VirtualMachine.html
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return nil, err
	}

	// Print summary per vm (see also: govc/vm/info.go)

	for _, vm := range vms {
		fmt.Printf("%s: %s\n", vm.Summary.Config.Name, vm.Summary.Config.GuestFullName)
		if vm.Summary.Config.Name == vmName {
			logger.Info(utils.Prettify(vm))
			return &vm, nil
		}
	}
	return nil, fmt.Errorf("Got none VM with vm-name:%s", vmName)
}

// =============== VM: Get VM Property ===============

// GetVMUUID .
func GetVMUUID(vm *object.VirtualMachine) string {
	ctx := context.Background()
	return vm.UUID(ctx)
}

// GetVMName .
func GetVMName(vm *object.VirtualMachine) string {
	var o mo.VirtualMachine
	ctx := context.Background()
	err := vm.Properties(ctx, vm.Reference(), []string{"config.name"}, &o)
	if err != nil {
		return ""
	}
	if o.Config != nil {
		return o.Config.Name
	}
	return ""
}

// =============== VM: poweroff/poweron/shutdown ===============

// WaitForVMPowerState .
func WaitForVMPowerState(vm *object.VirtualMachine, state types.VirtualMachinePowerState, tries int) error {
	ctx := context.Background()
	logger.Infof("Wait For VM %s ...", state)
	action := func(attempt uint) error {
		return vm.WaitForPowerState(ctx, types.VirtualMachinePowerStatePoweredOff)
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(tries)),
		strategy.Wait(20*time.Second),
	)
	curState, _ := vm.PowerState(ctx)
	logger.Infof("powerState(runtime/expected):%s/%s", curState, state)
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
	}
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
	}
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
