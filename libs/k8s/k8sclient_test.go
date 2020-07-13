package k8s

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	cfPath := "C:\\workspace\\go\\src\\platform\\kube\\10.25.119.71.config"
	client, err := NewClientWithRetry(cfPath)
	if err != nil {
		logger.Error(err.Error())
	}
	client.NameSpace = "vizion"

	execInput := ExecInput{
		PodName:       "cassandra-vset1-0",
		ContainerName: "cassandra",
		Command:       "/usr/bin/nodetool status | grep rack1",
	}
	output, err := client.Exec(execInput)
	logger.Info(output)
	logger.Info(err)
}
