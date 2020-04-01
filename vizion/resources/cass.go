package resources

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/op/go-logging"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

var logger = logging.MustGetLogger("test")

// GetS3UserRow ...
// u.Name is required
func (u *S3User) GetS3UserRow(session *gocql.Session) error {
	stmt, names := qb.Select("s3user").Where(qb.EqLit("name", fmt.Sprintf("'%s'", u.Name))).ToCql()
	f := gocqlx.Query(session.Query(stmt), names)
	logger.Infof("%+v", f)
	return f.Get(u)
}

// GetServiceRow ...
func (u *S3User) GetServiceRow(session *gocql.Session) error {
	stmt, names := qb.Select("s3user").Where(qb.EqLit("name", fmt.Sprintf("'%s'", u.Name))).ToCql()
	f := gocqlx.Query(session.Query(stmt), names)
	logger.Infof("%+v", f)
	return f.Get(u)
}
