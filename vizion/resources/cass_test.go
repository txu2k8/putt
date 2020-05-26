package resources

import (
	"pzatest/libs/db"
	"testing"

	"github.com/gocql/gocql"
)

func TestGetService(t *testing.T) {
	sessC := sessCluster{
		ConfigMap: map[string]db.CassConfig{
			"0": db.CassConfig{
				Hosts:    []string{"10.25.119.87"},
				Username: "caadmin",
				Password: "yjSJbEmPXmHfUbRa",
				Keyspace: "vizion",
				Port:     9042,
			},
		},
		SessionMap: map[string]*gocql.Session{"0": nil},
	}

	sessC.SetIndex("0")
	inputJSON := GetServiceInput{Type: 1024}
	services, _ := sessC.GetService(inputJSON)
	// logger.Infof("%+v", services)
	for _, sv := range services {
		logger.Infof(sv.IP)
	}

}
