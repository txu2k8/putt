package resources

import (
	"testing"
)

func TestGetService(t *testing.T) {
	sessC := sessCluster{
		ConfigMap: map[string]CassConfig{
			"0": CassConfig{
				IPs:      []string{"10.25.119.87"},
				User:     "caadmin",
				Password: "yjSJbEmPXmHfUbRa",
				Keyspace: "vizion",
				Port:     9042,
			},
		},
	}

	sessC.SetIndex("0")
	inputJSON := GetServiceInput{Type: 1024}
	services, _ := sessC.GetService(inputJSON)
	// logger.Infof("%+v", services)
	for _, sv := range services {
		logger.Infof(sv.IP)
	}

}
