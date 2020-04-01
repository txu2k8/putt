package db

import (
	"fmt"
	_ "gtest/testinit"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

// s3User ...
type s3User struct {
	Name           string            `db:"name"`
	Bucket         []string          `db:"bucket"`
	BucketGroup    []string          `db:"bucketgroup"`
	CTime          time.Time         `db:"c_time"`
	ChangePassword bool              `db:"change_password"`
	Group          []string          `db:"group"`
	Info           string            `db:"info"`
	MTime          time.Time         `db:"m_time"`
	Password       string            `db:"password"`
	PasswordMTime  time.Time         `db:"password_m_time"`
	S3Access       map[string]string `db:"s3access"`
	Status         int               `db:"status"`
	Tenant         string            `db:"tenant"`
}

// getS3UserRow ...
func (u *s3User) getS3UserRow(session *gocql.Session) error {
	stmt, names := qb.Select("s3user").Where(qb.EqLit("name", fmt.Sprintf("'%s'", u.Name))).ToCql()
	f := gocqlx.Query(session.Query(stmt), names)
	logger.Infof("%+v", f)
	return f.Get(u)
}

// getS3UserRow ...
func (u *s3User) getS3UserRow2(session *gocql.Session) error {
	stmt, names := qb.Select("s3user").AllowFiltering().ToCql()
	f := gocqlx.Query(session.Query(stmt), names)
	logger.Infof("%+v", f)
	return f.Get(u)
}

func TestDB(t *testing.T) {
	cassConfig := CassConfig{
		host:     "10.25.119.84",
		user:     "caadmin",
		pwd:      "YFPliyZsejloVVrU",
		keyspace: "vizion",
		port:     9042,
	}
	session, err := NewSessionWithRetry(&cassConfig)
	if err != nil {
		logger.Panic(err)
	}

	s3user := s3User{}
	s3user.getS3UserRow2(session)
	logger.Info(s3user)
}
