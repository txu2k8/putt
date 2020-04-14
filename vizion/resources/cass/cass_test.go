package cass

import (
	"gtest/libs/db"
	"testing"
)

func TestGetService(t *testing.T) {
	masterConfig := db.CassConfig{
		Hosts:    "10.25.119.84",
		Username: "caadmin",
		Password: "YFPliyZsejloVVrU",
		Keyspace: "vizion",
		Port:     9042,
	}
	session, _ := db.NewSessionWithRetry(&masterConfig)

	// inputJSON := GetServiceInput{Type: 1024}
	// services, _ := GetService(session, inputJSON)
	// // logger.Infof("%+v", services)
	// for _, sv := range services {
	// 	logger.Infof(sv.IP)
	// }

	// s3users, _ := GetS3User(session)
	// for _, s3user := range s3users {
	// 	logger.Info(s3user.Name)
	// }

	// vols, _ := GetVolume(session)
	// for _, vol := range vols {
	// 	logger.Info(vol.Name)
	// }

	DeleteFromTable(session, "s3user")

}
