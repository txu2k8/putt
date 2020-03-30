package config

import (
	"os"

	"github.com/jinzhu/configor"
)

// define const values
const (
	LogLevel = 3
	Root     = os.Getenv("PWD")
)

// Config application all configs
var Config = struct {
	Port  uint `default:"8000" env:"PORT"`
	Pprof bool `default:"false" env:"PPROF"`
	ES    struct {
		URLs                string `env:"URLs" default:"http://127.0.0.1:9200"`
		Username            string `env:"Username" default:"root"`
		Password            string `env:"Password" default:"password"`
		HealthcheckInterval int    `env:"HealthcheckInterval" default:"20"`
	}
	MonitorES struct {
		URLs                string `env:"URLs" default:"http://10.180.113.73:9211"`
		Username            string `env:"Username" default:"root"`
		Password            string `env:"Password" default:"password"`
		HealthcheckInterval int    `env:"HealthcheckInterval" default:"20"`
	}
	Cassandra struct {
		Cluster       string `env:"Cluster" default:"10.203.79.240"`
		Keyspace      string `env:"Keyspace" default:"vizion"`
		Username      string `env:"CASUser" default:"caadmin"`
		Password      string `env:"CASPwd" default:"nSHduPhOCfYojtRX"`
		Timeout       int    `env:"CASTimeout" default:"6000"`
		CacheInterval int    `env:"CacheInterval" default:"300"`
	}
	Master struct {
		MaxWatcherCount          int    `env:"MaxWatcherCount" default:"50"`
		MaxNotRefreshCount       int    `env:"MaxNotRefreshCount" default:"5"`
		MaxIndexerCount          int    `env:"MaxIndexerCount" default:"10000"`
		ChannelSize              int    `env:"ChannelSize" default:"100"`
		WatchInterval            int    `env:"WatchInterval" default:"300"`
		WatchPageSize            int64  `env:"WatchPageSize" default:"500"`
		IndexMode                string `env:"INDEX_MODE" default:"limited"`
		SnapshotSchedule         string `env:"SNAP_SCHEDULE" default:"0 0 0,6,12,18 * * *"`
		SnapshotDisable          string `env:"SNAP_DISABLE" default:"false"`
		WaitIndexTimeoutDuration int    `env:"WaitIndexTimeoutDuration" default:"60"`
	}
	Worker struct {
		MaxNum     int    `env:"MAX_WORKER_NUM" default:"15"`
		VIPTenants string `env:"VIP_TENANTS"`
		V2Style    string `env:"V2_STYLE" default:"true"`
		Merge      struct {
			Enabled  string `env:"MERGE_ENABLED" default:"true"`
			PoolSize int    `env:"MERGE_POOL_SIZE" default:"100"`
		}
	}
	Locale  string `env:"Locale" default:"en-US"`
	Version string `env:"Version" default:"v1.0.0"`
}{}

func init() {
	if err := configor.Load(&Config); err != nil {
		panic(err)
	}
}
