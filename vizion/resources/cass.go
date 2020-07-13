package resources

import (
	"errors"
	"fmt"
	"platform/config"
	"platform/libs/db/cql"

	"github.com/chenhg5/collection"
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
	GetServiceByTypeVsetID(serviceType int, vsetID int) ([]Service, error)
	GetAllServices(serviceType int) (svArr []Service)
	MjcachedplServices() (svArr []Service)
	DjcachedplServices() (svArr []Service)
	AnchordplServices() (svArr []Service)
	JddplServices() (svArr []Service)
	ServicedplServices() (svArr []Service)
	FlushdplServices() (svArr []Service)
	McmapdplServices() (svArr []Service)
	DcmapdplServices() (svArr []Service)
	CmapdplServices() (svArr []Service)
	Vizions3Services() (svArr []Service)
	DpldagentServices() (svArr []Service)
	GetVolume() ([]Volume, error)
	GetTenant() ([]Tenant, error)
	GetS3User() ([]S3User, error)
	GetS3Bucket() ([]S3Bucket, error)
	GetS3BucketGroup() ([]S3BucketGroup, error)
	Execute(cmd string) error
	TruncateTable(table string) error
}

type sessCluster struct {
	Session    *gocql.Session
	ConfigMap  map[string]cql.CassConfig // {"0": cql.CassConfig}
	SessionMap map[string]*gocql.Session // {"0": *gocql.Session}
	VsetIDs    []int                     // base.VsetIDs
}

func newSessCluster(v *Vizion) *sessCluster {
	return &sessCluster{
		ConfigMap:  v.GetCassConfig(),
		SessionMap: make(map[string]*gocql.Session, len(v.Base.VsetIDs)+1),
		VsetIDs:    v.Base.VsetIDs,
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

	cqlConfig := c.ConfigMap[index]
	session, _ := cql.NewSessionWithRetry(&cqlConfig)
	c.SessionMap[index] = session
	c.Session = c.SessionMap[index]
	return c
}

// Execute ...
func (c *sessCluster) Execute(cmd string) error {
	logger.Info(cmd)
	if err := c.Session.Query(cmd).Exec(); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

// DeleteFromTable ... TODO
func (c *sessCluster) DeleteFromTable(table string) {
	stmt, _ := qb.Delete(table).Where(qb.EqLit("name", fmt.Sprintf("%s", "vset1_s3user"))).ToCql()
	logger.Info(stmt)
}

// TruncateTable ...
func (c *sessCluster) TruncateTable(table string) error {
	return c.Execute("TRUNCATE " + table)
}

// =============== select from table ===============

// GetCassandraCluster ...
func (c *sessCluster) GetCassandraCluster() ([]CassandraCluster, error) {
	var ccs []CassandraCluster
	stmt, names := qb.Select("cassandra_cluster").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&ccs)
	return ccs, err
}

// GetNode ...
func (c *sessCluster) GetNode() ([]Node, error) {
	var nodes []Node
	stmt, names := qb.Select("node").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&nodes)
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
	return qb.Select("service").Where(qb.EqLit("type", fmt.Sprintf("%d", serviceType))).AllowFiltering().ToCql()

}

// SelectServiceByTypeID ...
func SelectServiceByTypeID(serviceType int, serviceUUID string) (stmt string, names []string) {
	return qb.Select("service").Where(qb.EqLit("type", fmt.Sprintf("%d", serviceType))).Where(qb.EqLit("id", serviceUUID)).AllowFiltering().ToCql()
}

// SelectServiceByTypeVsetID ...
func SelectServiceByTypeVsetID(serviceType int, vsetID int) (stmt string, names []string) {
	return qb.Select("service").Where(qb.EqLit("type", fmt.Sprintf("%d", serviceType))).Where(qb.EqLit("vset_id", fmt.Sprintf("%d", vsetID))).AllowFiltering().ToCql()
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
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&services)
	return services, err
}

// GetServiceByType ...
func (c *sessCluster) GetServiceByType(serviceType int) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByType(serviceType)
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&services)
	return services, err
}

// GetServiceByTypeID ...
func (c *sessCluster) GetServiceByTypeID(serviceType int, serviceUUID string) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByTypeID(serviceType, serviceUUID)
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&services)
	return services, err
}

// GetServiceByTypeID ...
func (c *sessCluster) GetServiceByTypeVsetID(serviceType int, vsetID int) ([]Service, error) {
	var services []Service
	stmt, names := SelectServiceByTypeVsetID(serviceType, vsetID)
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&services)
	return services, err
}

// GetServices .
func (c *sessCluster) GetAllServices(serviceType int) (svArr []Service) {
	vset0SvTypeArr := []int{
		config.Jddpl.Type,
		config.Dcmapdpl.Type,
		config.Mcmapdpl.Type,
		config.Dpldagent.Type,
	}

	if collection.Collect(vset0SvTypeArr).Contains(serviceType) {
		vsetSvArr, _ := c.GetServiceByTypeVsetID(serviceType, 0)
		svArr = append(svArr, vsetSvArr...)

	} else {
		for _, vsetID := range c.VsetIDs {
			vsetSvArr, _ := c.GetServiceByTypeVsetID(serviceType, vsetID)
			svArr = append(svArr, vsetSvArr...)
		}
	}

	return
}

// MjcachedplServices .
func (c *sessCluster) MjcachedplServices() (svArr []Service) {
	return c.GetAllServices(config.Mjcachedpl.Type)
}

// DjcachedplServices .
func (c *sessCluster) DjcachedplServices() (svArr []Service) {
	return c.GetAllServices(config.Djcachedpl.Type)
}

// AnchordplServices .
func (c *sessCluster) AnchordplServices() (svArr []Service) {
	svArr = append(svArr, c.MjcachedplServices()...)
	svArr = append(svArr, c.DjcachedplServices()...)
	return
}

// JddplServices .
func (c *sessCluster) JddplServices() (svArr []Service) {
	return c.GetAllServices(config.Jddpl.Type)
}

// ServicedplServices .
func (c *sessCluster) ServicedplServices() (svArr []Service) {
	return c.GetAllServices(config.Servicedpl.Type)
}

// FlushdplServices .
func (c *sessCluster) FlushdplServices() (svArr []Service) {
	return c.GetAllServices(config.Flushdpl.Type)
}

// McmapdplServices .
func (c *sessCluster) McmapdplServices() (svArr []Service) {
	return c.GetAllServices(config.Mcmapdpl.Type)
}

// DcmapdplServices .
func (c *sessCluster) DcmapdplServices() (svArr []Service) {
	return c.GetAllServices(config.Dcmapdpl.Type)
}

// CmapdplServices .
func (c *sessCluster) CmapdplServices() (svArr []Service) {
	svArr = append(svArr, c.McmapdplServices()...)
	svArr = append(svArr, c.DcmapdplServices()...)
	return
}

// Vizions3Services .
func (c *sessCluster) Vizions3Services() (svArr []Service) {
	return c.GetAllServices(config.Vizions3.Type)
}

// DpldagentServices .
func (c *sessCluster) DpldagentServices() (svArr []Service) {
	return c.GetAllServices(config.Dpldagent.Type)
}

// GetVolume ...
func (c *sessCluster) GetVolume() ([]Volume, error) {
	var volumes []Volume
	stmt, names := qb.Select("volume").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&volumes)
	return volumes, err
}

// GetTenant ...
func (c *sessCluster) GetTenant() ([]Tenant, error) {
	var tenants []Tenant
	stmt, names := qb.Select("tenant").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&tenants)
	return tenants, err
}

// GetS3User ...
func (c *sessCluster) GetS3User() ([]S3User, error) {
	var s3Users []S3User
	stmt, names := qb.Select("s3user").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&s3Users)
	return s3Users, err
}

// GetS3Bucket ...
func (c *sessCluster) GetS3Bucket() ([]S3Bucket, error) {
	var s3buckets []S3Bucket
	stmt, names := qb.Select("s3bucket").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&s3buckets)
	return s3buckets, err
}

// GetS3BucketGroup ...
func (c *sessCluster) GetS3BucketGroup() ([]S3BucketGroup, error) {
	var s3bgs []S3BucketGroup
	stmt, names := qb.Select("s3bucketgroup").ToCql()
	f := gocqlx.Query(c.Session.Query(stmt), names)
	logger.Debugf("%v", f)
	err := f.SelectRelease(&s3bgs)
	return s3bgs, err
}
