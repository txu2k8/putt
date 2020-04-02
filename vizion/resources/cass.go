package resources

import (
	"errors"
	"fmt"
	"gtest/libs/db"

	"github.com/gocql/gocql"
	"github.com/op/go-logging"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

var logger = logging.MustGetLogger("test")

// CassCluster ...
type CassCluster struct {
	Session    *gocql.Session
	sessionMap map[string]*gocql.Session
	MdSession  *gocql.Session
}

// cass clusters ...
var masterCasCluster *gocql.ClusterConfig

// SetIndex ...
func (c *CassCluster) SetIndex(index string) {
	c.MdSession = c.sessionMap[index]
}

// InitSessionMap ...
func (c *CassCluster) InitSessionMap(masterConfig *db.CassConfig) {
	masterSession, e := db.NewSessionWithRetry(masterConfig)
	if e != nil {
		panic(e)
	}
	c.Session = masterSession

	var master = make(map[string]*gocql.Session)
	master["0"] = masterSession
	c.sessionMap = master
}

// =============== select from table ===============

// GetCassandraCluster ...
func GetCassandraCluster(session *gocql.Session) ([]CassandraCluster, error) {
	var ccs []CassandraCluster
	stmt, names := qb.Select("cassandra_cluster").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&ccs)
	return ccs, err
}

// GetNode ...
func GetNode(session *gocql.Session) ([]Node, error) {
	var nodes []Node
	stmt, names := qb.Select("node").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&nodes)
	return nodes, err
}

// GetServiceInput ...
type GetServiceInput struct {
	Type int
	ID   string
}

// SelectService ...
func SelectService() (stmt string, names []string) {
	return qb.Select("service").ToCql()
}

// SelectServiceByType ...
func SelectServiceByType(serviceType int) (stmt string, names []string) {
	return qb.Select("service").Where(qb.EqLit("type", fmt.Sprintf("%d", serviceType))).ToCql()

}

// SelectServiceByTypeID ...
func SelectServiceByTypeID(serviceType int, serviceUUID string) (stmt string, names []string) {
	return qb.Select("service").Where(qb.EqLit("type", fmt.Sprintf("%d", serviceType))).Where(qb.EqLit("id", serviceUUID)).ToCql()
}

// GetService ...
func GetService(session *gocql.Session, inputJSON GetServiceInput) ([]Service, error) {
	var (
		stmt     string
		names    []string
		services []Service
	)

	switch {
	case inputJSON.Type == 0 && inputJSON.ID == "":
		stmt, names = SelectService()
	case inputJSON.ID != "" && inputJSON.Type != 0:
		stmt, names = SelectServiceByTypeID(inputJSON.Type, inputJSON.ID)
	case inputJSON.Type != 0:
		stmt, names = SelectServiceByType(inputJSON.Type)
	default:
		return services, errors.New("Error type or id")
	}
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&services)
	return services, err
}

// GetServiceByType ...
func GetServiceByType(session *gocql.Session, serviceType int) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByType(serviceType)
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&services)
	return services, err
}

// GetServiceByTypeID ...
func GetServiceByTypeID(session *gocql.Session, serviceType int, serviceUUID string) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByTypeID(serviceType, serviceUUID)
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&services)
	return services, err
}

// GetVolume ...
func GetVolume(session *gocql.Session) ([]Volume, error) {
	var volumes []Volume
	stmt, names := qb.Select("volume").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&volumes)
	return volumes, err
}

// GetTenant ...
func GetTenant(session *gocql.Session) ([]Tenant, error) {
	var tenants []Tenant
	stmt, names := qb.Select("tenant").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&tenants)
	return tenants, err
}

// GetS3User ...
func GetS3User(session *gocql.Session) ([]S3User, error) {
	var s3Users []S3User
	stmt, names := qb.Select("s3user").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&s3Users)
	return s3Users, err
}

// GetS3Bucket ...
func GetS3Bucket(session *gocql.Session) ([]S3Bucket, error) {
	var s3buckets []S3Bucket
	stmt, names := qb.Select("s3bucket").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&s3buckets)
	return s3buckets, err
}

// GetS3BucketGroup ...
func GetS3BucketGroup(session *gocql.Session) ([]S3BucketGroup, error) {
	var s3bgs []S3BucketGroup
	stmt, names := qb.Select("s3bucketgroup").ToCql()
	err := gocqlx.Query(session.Query(stmt), names).SelectRelease(&s3bgs)
	return s3bgs, err
}
