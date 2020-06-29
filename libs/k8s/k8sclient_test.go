package k8s

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	cfPath := "C:\\workspace\\config"
	client, err := NewClientWithRetry(cfPath)
	if err != nil {
		logger.Error(err.Error())
	}
	client.NameSpace = "vizion"

	client.GetStatefulSetsNameArrByLabel("name=servicedpl-1-1")
}
