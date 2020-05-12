package db

import (
	"fmt"
	_ "pzatest/testinit"
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
func getS3UserRow(session *gocql.Session, s3userName string) ([]s3User, error) {
	var s3users []s3User
	stmt, names := qb.Select("s3user").Where(qb.EqLit("name", fmt.Sprintf("'%s'", s3userName))).ToCql()
	f := gocqlx.Query(session.Query(stmt), names)
	logger.Infof("%+v", f)
	// return f.Get(u)
	e := f.SelectRelease(&s3users)
	return s3users, e
}

func TestDB(t *testing.T) {
	cassConfig := CassConfig{
		Hosts:    "10.25.119.84",
		Username: "caadmin",
		Password: "YFPliyZsejloVVrU",
		Keyspace: "vizion",
		Port:     9042,
	}
	session, err := NewSessionWithRetry(&cassConfig)
	if err != nil {
		logger.Panic(err)
	}

	s3users, _ := getS3UserRow(session, "vset1_s3user")
	logger.Infof("%+v", s3users)
}
