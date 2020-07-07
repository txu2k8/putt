package resources

import (
	"encoding/json"
	"putt/libs/db/etcd"
	"putt/libs/utils"
	"strings"
)

// EtcdGetter has a method to return a EtcdInterface.
// A client should implement this interface.
type EtcdGetter interface {
	Etcd(host string) EtcdInterface
}

// EtcdInterface has methods to work on Etcd resources.
type EtcdInterface interface {
	GetStgUnitArr() (stgArr []StgUnit, err error)
	GetStgUnitNumber() (number int64, err error)
}

type etcdv3 struct {
	*etcd.Client
}

func newEtcd(v *Vizion) *etcdv3 {
	certFile, keyFile, trustedCAFile := "", "", ""
	for _, c := range v.GetEtcdConfig() {
		switch {
		case strings.Contains(c, "peer.crt"):
			certFile = c
		case strings.Contains(c, "peer.key"):
			keyFile = c
		case strings.Contains(c, "ca.crt"):
			trustedCAFile = c
		default:
			continue
		}
	}
	cfg := etcd.Config{
		Endpoints:     v.MasterNode().GetEtcdEndpoints(),
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: trustedCAFile,
	}
	logger.Debugf(utils.Prettify(cfg))
	cli, _ := etcd.NewClientWithRetry(cfg)
	return &etcdv3{cli}
}

// StgUnit in etcd
type StgUnit struct {
	ServiceUUID string `json:"service_uuid"`
	DiskPname   string `json:"disk_pname"`
	UnitUUID    string `json:"unit_uuid"`
	Next        string `json:"next"`
	Pair        string `json:"pair"`
	Offset      int64  `json:"offset"`
	Status      int    `json:"status"`
	Size        int64  `json:"size"`
	VsetID      int    `json:"vset_id"`
	Idx         int    `json:"idx"`
	JDGroupID   int    `json:"jd_group_id"`
	DiskID      int    `json:"disk_id"`
}

func (e *etcdv3) GetStgUnitArr() (stgArr []StgUnit, err error) {
	resp, err := e.GetPrefix("/vizion/dpl/stg_unit")
	if err != nil {
		return
	}
	for _, kv := range resp.Kvs {
		su := StgUnit{}
		json.Unmarshal(kv.Value, &su)
		// logger.Info(utils.Prettify(su))}
		stgArr = append(stgArr, su)
	}
	return
}

func (e *etcdv3) GetStgUnitNumber() (number int64, err error) {
	resp, err := e.GetPrefix("/vizion/dpl/stg_unit")
	if err != nil {
		return
	}
	number = resp.Count
	return
}
