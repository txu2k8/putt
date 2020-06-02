package resources

import (
	"errors"
	"fmt"
	"pzatest/libs/db"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

// CassClusterGetter has a method to return a CassCluster.
type CassClusterGetter interface {
	Cass() CassCluster
}

// CassCluster ...
type CassCluster interface {
	SetIndex(index string) CassCluster
	GetCassandraCluster() ([]CassandraCluster, error)
	GetNode() ([]Node, error)
	GetService(inputJSON GetServiceInput) ([]Service, error)
	GetServiceByType(serviceType int) ([]Service, error)
	GetServiceByTypeID(serviceType int, serviceUUID string) ([]Service, error)
	GetVolume() ([]Volume, error)
	GetTenant() ([]Tenant, error)
	GetS3User() ([]S3User, error)
	GetS3Bucket() ([]S3Bucket, error)
	GetS3BucketGroup() ([]S3BucketGroup, error)
	TruncateTable(table string) error
}

type sessCluster struct {
	Session    *gocql.Session
	ConfigMap  map[string]db.CassConfig  // {"0": db.CassConfig}
	SessionMap map[string]*gocql.Session // {"0": *gocql.Session}
}

func newSessCluster(v *Vizion) *sessCluster {
	return &sessCluster{
		ConfigMap:  v.GetCassConfig(),
		SessionMap: map[string]*gocql.Session{"0": nil},
	}
}

// SetIndex ...
func (c *sessCluster) SetIndex(index string) CassCluster {
	if _, ok := c.SessionMap[index]; ok {
		if c.SessionMap[index] != nil {
			c.Session = c.SessionMap[index]
			return c
		}
	}

	dbConfig := c.ConfigMap[index]
	session, _ := db.NewSessionWithRetry(&dbConfig)
	c.SessionMap[index] = session
	c.Session = c.SessionMap[index]
	return c
}

// DeleteFromTable ... TODO
func (c *sessCluster) DeleteFromTable(table string) {
	stmt, _ := qb.Delete(table).Where(qb.EqLit("name", fmt.Sprintf("%s", "vset1_s3user"))).ToCql()
	logger.Info(stmt)
}

// TruncateTable ... TODO
func (c *sessCluster) TruncateTable(table string) error {
	stmt, _ := qb.Delete(table).Where(qb.EqLit("name", fmt.Sprintf("%s", "vset1_s3user"))).ToCql()
	logger.Info(stmt)
	return nil
}

// =============== select from table ===============

// GetCassandraCluster ...
func (c *sessCluster) GetCassandraCluster() ([]CassandraCluster, error) {
	var ccs []CassandraCluster
	stmt, names := qb.Select("cassandra_cluster").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&ccs)
	return ccs, err
}

// GetNode ...
func (c *sessCluster) GetNode() ([]Node, error) {
	var nodes []Node
	stmt, names := qb.Select("node").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&nodes)
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
func (c *sessCluster) GetService(inputJSON GetServiceInput) ([]Service, error) {
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
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&services)
	return services, err
}

// GetServiceByType ...
func (c *sessCluster) GetServiceByType(serviceType int) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByType(serviceType)
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&services)
	return services, err
}

// GetServiceByTypeID ...
func (c *sessCluster) GetServiceByTypeID(serviceType int, serviceUUID string) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByTypeID(serviceType, serviceUUID)
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&services)
	return services, err
}

// GetVolume ...
func (c *sessCluster) GetVolume() ([]Volume, error) {
	var volumes []Volume
	stmt, names := qb.Select("volume").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&volumes)
	return volumes, err
}

// GetTenant ...
func (c *sessCluster) GetTenant() ([]Tenant, error) {
	var tenants []Tenant
	stmt, names := qb.Select("tenant").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&tenants)
	return tenants, err
}

// GetS3User ...
func (c *sessCluster) GetS3User() ([]S3User, error) {
	var s3Users []S3User
	stmt, names := qb.Select("s3user").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&s3Users)
	return s3Users, err
}

// GetS3Bucket ...
func (c *sessCluster) GetS3Bucket() ([]S3Bucket, error) {
	var s3buckets []S3Bucket
	stmt, names := qb.Select("s3bucket").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&s3buckets)
	return s3buckets, err
}

// GetS3BucketGroup ...
func (c *sessCluster) GetS3BucketGroup() ([]S3BucketGroup, error) {
	var s3bgs []S3BucketGroup
	stmt, names := qb.Select("s3bucketgroup").ToCql()
	err := gocqlx.Query(c.Session.Query(stmt), names).SelectRelease(&s3bgs)
	return s3bgs, err
}
