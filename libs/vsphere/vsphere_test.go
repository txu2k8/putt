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
	// c.GetVMDetails("Ceph-61")
	vm, _ := c.GetVMByName("Ceph-61")
	vmName := GetVMName(vm)
	logger.Info(vmName)
	// PowerOnVM(vm)
}
