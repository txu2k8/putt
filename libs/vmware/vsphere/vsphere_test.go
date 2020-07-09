package vsphere

import (
	"testing"
)

func TestGetVM(t *testing.T) {
	vc := VspConfig{
		Host:     "10.25.1.8",
		Username: "stress@panzura.com",
		Password: "P@ssword1",
	}

	c, _ := NewClientWithRetry(&vc)
	// vmDteail, _ := c.GetVMDetails("WIN2012R2-txu-HQAD-21")
	// logger.Info(utils.Prettify(vmDteail))

	vm, _ := c.GetVMByName("WIN2012R2-txu-HQAD-21")
	// vm, _ := c.GetVMByUUID("423a8b41-ea41-eae9-564a-2ecd07ec81ba")
	// PowerOptVM(vm, "poweron")
	// MultiPowerOptVM([]OptVM{{Opt: "shutdown", VM: vm}})

	ip := GetVMName(vm)
	logger.Info(ip)
}
