package config

import (
	"fmt"
	"putt/libs/utils"
	"putt/types"
	"strings"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// ========== const: etcd Settings ==========
const (
	EtcdPort     = 2379
	EtcdCertPath = "/etc/kubernetes/pki/etcd/"
)

// ========== const: K8S Settings ==========
const (
	K8sDeployment   = "deployment"
	K8sStatefulsets = "statefulsets"
	K8sDaemonsets   = "daemonsets"
)

// ========== const: Upgrade Image Settings ==========
const (
	DplBuildIP          = "10.199.116.1"
	DplBuildPath        = "/home/project/dpl/develop/dpl"
	DplBuildLocalPath   = "/mnt/vizion/QA/dpl/"
	RemoteRegistryIP    = "10.180.1.45"
	RemoteRegistry      = "registry.vizion.ai"
	RemoteDplRegistry   = "registry.vizion.ai/stable/dpl"
	LocalRegistry       = "registry.vizion.local"
	LocalDplRegistry    = "registry.vizion.local/stable/dpl"
	DefaultDplImage     = "registry.vizion.local/library/dpl:tmp"
	DplmanagerLocalPath = "/usr/bin/dplmanager"
	JDevicePath         = "/dev/j_device"
)

// ========== const: DPL Settings ==========
const (
	// DPL Channel type
	CHTYPEBD = "CH_TYPE_BD"
	CHTYPES3 = "CH_TYPE_S3"
)

// Service define the service information
type Service struct {
	Name       string   // service name
	Path       string   // service binary path in pod
	GitPath    string   // servicepath in gitlab
	Type       int      // service Type ID
	TypeName   string   // service type name
	NameSpace  string   // service in k8s namespace
	K8sKind    string   // service resource kind in k8s, sts,ds,deploy...
	PodLabel   string   // pod label template
	NodeLabel  string   // node label template
	Container  string   // service pod container name
	Replicas   int      // service replicas
	GetPid     string   // the cmd to get service pid, eg: ps -ax | grep dpl
	LogPathArr []string // the service log path list
}

// CleanItem define the cleanup item
type CleanItem struct {
	Name string
	Arg  []string
}

// GetPodLabel .
func (sv *Service) GetPodLabel(base types.BaseInput) (podLabel string) {
	keyValue := strings.Split(sv.PodLabel, "=")
	labelKey := keyValue[0]
	labelValue := keyValue[1]

	var podLabelValueArr []string
	var podLabelValueStr string
	switch sv.Type {
	case MasterCass.Type, MysqlCluster.Type, MysqlOperator.Type, MysqlRouter.Type, ETCD.Type,
		Dplexporter.Type, Dplmanager.Type, Nfsprovisioner.Type:
		podLabelValueArr = []string{labelValue} // fixed label key=Value
	case SubCass.Type:
		for _, vsetID := range base.VsetIDs {
			vsetPodValue := fmt.Sprintf("%s%d", labelValue, vsetID)
			podLabelValueArr = append(podLabelValueArr, vsetPodValue)
		}
	case Jddpl.Type:
		podLabelValueArr = []string{labelValue}
		for _, jdGroupID := range base.JDGroupIDs {
			jdLabelValue := fmt.Sprintf("%s-%d", labelValue, jdGroupID)
			podLabelValueArr = append(podLabelValueArr, jdLabelValue)
		}
	case Servicedpl.Type:
		podLabelValueArr = []string{labelValue}
		for _, vsetID := range base.VsetIDs {
			vsetPodValue := fmt.Sprintf("%s-%d", labelValue, vsetID)
			podLabelValueArr = append(podLabelValueArr, vsetPodValue)
			for _, dplGroupID := range base.DPLGroupIDs {
				dplLabelValue := fmt.Sprintf("%s-%d-%d", labelValue, vsetID, dplGroupID)
				podLabelValueArr = append(podLabelValueArr, dplLabelValue)
			}
		}
	case Mjcachedpl.Type, Djcachedpl.Type:
		podLabelValueArr = []string{labelValue}
		for _, vsetID := range base.VsetIDs {
			vsetPodValue := fmt.Sprintf("%s-%d", labelValue, vsetID)
			podLabelValueArr = append(podLabelValueArr, vsetPodValue)
			for _, jcacheGroupID := range base.JcacheGroupIDs {
				jcacheLabelValue := fmt.Sprintf("%s-%d-%d", labelValue, vsetID, jcacheGroupID)
				podLabelValueArr = append(podLabelValueArr, jcacheLabelValue)
			}
		}
	case Mcmapdpl.Type, Dcmapdpl.Type:
		podLabelValueArr = []string{labelValue}
		for _, cmapGroupID := range base.CmapGroupIDs {
			cmapLabelValue := fmt.Sprintf("%s-%d", labelValue, cmapGroupID)
			podLabelValueArr = append(podLabelValueArr, cmapLabelValue)
		}
	case ES.Type:
		podLabelValueArr = []string{labelValue}
		for _, vsetID := range base.VsetIDs {
			vsetPodValue := fmt.Sprintf("%s-%d", labelValue, vsetID)
			podLabelValueArr = append(podLabelValueArr, vsetPodValue)
		}

	default:
		for _, vsetID := range base.VsetIDs {
			vsetPodValue := fmt.Sprintf("%s-%d", labelValue, vsetID)
			podLabelValueArr = append(podLabelValueArr, vsetPodValue)
		}
	}
	podLabelValueStr = strings.Join(podLabelValueArr, ",")
	podLabel = fmt.Sprintf("%s in (%s)", labelKey, podLabelValueStr)
	logger.Debugf("PodLabel: %s", utils.Prettify(podLabel))
	return
}

// GetNodeLabelArr .
func (sv *Service) GetNodeLabelArr(base types.BaseInput) (nodeLabelKeyArr, nodeLabelKeyValueArr []string) {
	keyValue := strings.Split(sv.NodeLabel, "=")
	labelKey := keyValue[0]
	labelValue := keyValue[1]

	switch sv.Type {
	case Dplexporter.Type, Dplmanager.Type, MysqlCluster.Type, ETCD.Type,
		Cdcgcbd.Type, Cdcgcs3.Type, Nfsprovisioner.Type:
		nodeLabelKeyArr = []string{labelKey}
		nodeLabelKeyValueArr = []string{labelKey + "=" + labelValue}
	case Jddpl.Type:
		for _, jdGroupID := range base.JDGroupIDs {
			labelK := fmt.Sprintf("%s-%d", labelKey, jdGroupID)
			labelKV := fmt.Sprintf("%s=%s", labelK, labelValue)
			nodeLabelKeyArr = append(nodeLabelKeyArr, labelK)
			nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, labelKV)
		}
	case Servicedpl.Type:
		for _, vsetID := range base.VsetIDs {
			vsetLabelK := fmt.Sprintf("%s-%d", labelKey, vsetID)
			vsetLabelKV := fmt.Sprintf("%s=%s", vsetLabelK, labelValue)
			nodeLabelKeyArr = append(nodeLabelKeyArr, vsetLabelK)
			nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, vsetLabelKV)
			for _, dplGroupID := range base.DPLGroupIDs {
				labelK := fmt.Sprintf("%s-%d-%d", labelKey, vsetID, dplGroupID)
				labelKV := fmt.Sprintf("%s=%s", labelK, labelValue)
				nodeLabelKeyArr = append(nodeLabelKeyArr, labelK)
				nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, labelKV)
			}
		}
	case Mjcachedpl.Type, Djcachedpl.Type:
		for _, vsetID := range base.VsetIDs {
			vsetLabelK := fmt.Sprintf("%s-%d", labelKey, vsetID)
			vsetLabelKV := fmt.Sprintf("%s=%s", vsetLabelK, labelValue)
			nodeLabelKeyArr = append(nodeLabelKeyArr, vsetLabelK)
			nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, vsetLabelKV)
			for _, jcacheGroupID := range base.DPLGroupIDs {
				labelK := fmt.Sprintf("%s-%d-%d", labelKey, vsetID, jcacheGroupID)
				labelKV := fmt.Sprintf("%s=%s", labelK, labelValue)
				nodeLabelKeyArr = append(nodeLabelKeyArr, labelK)
				nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, labelKV)
			}
		}
	case Mcmapdpl.Type, Dcmapdpl.Type:
		for _, cmapGroupID := range base.CmapGroupIDs {
			labelK := fmt.Sprintf("%s-%d", labelKey, cmapGroupID)
			labelKV := fmt.Sprintf("%s=%s", labelK, labelValue)
			nodeLabelKeyArr = append(nodeLabelKeyArr, labelK)
			nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, labelKV)
		}
	case ES.Type:
		// ES need bind to bd-vset nodes
		bdKV := strings.Split(Dpldagent.NodeLabel, "=")
		for _, vsetID := range base.VsetIDs {
			vsetBdK := fmt.Sprintf("%s-%d", bdKV[0], vsetID)
			vsetBdKV := fmt.Sprintf("%s=%s", vsetBdK, bdKV[1])
			nodeLabelKeyArr = []string{fmt.Sprintf("%s,%s", vsetBdK, labelKey)}
			nodeLabelKeyValueArr = []string{fmt.Sprintf("%s,%s", vsetBdKV, labelKey+"="+labelValue)}
		}
	default:
		for _, vsetID := range base.VsetIDs {
			vsetLabelK := fmt.Sprintf("%s-%d", labelKey, vsetID)
			vsetLabelKV := fmt.Sprintf("%s=%s", vsetLabelK, labelValue)
			nodeLabelKeyArr = append(nodeLabelKeyArr, vsetLabelK)
			nodeLabelKeyValueArr = append(nodeLabelKeyValueArr, vsetLabelKV)
		}
	}

	// logger.Debug(utils.Prettify(nodeLabelKeyArr))
	logger.Debugf("NodeLabel: %s", utils.Prettify(nodeLabelKeyValueArr))
	return
}

// GetLogDirArr .
func (sv *Service) GetLogDirArr(base types.BaseInput) (logDirArr []string) {
	switch sv.Type {
	case Jddpl.Type: // *-1
		for _, jdGroupID := range base.JDGroupIDs {
			for _, logPath := range sv.LogPathArr {
				jdLogPath := fmt.Sprintf("%s-%d", logPath, jdGroupID)
				logDirArr = append(logDirArr, jdLogPath)
			}
		}
	case Servicedpl.Type: // *-vset1-1
		for _, vsetID := range base.VsetIDs {
			for _, logPath := range sv.LogPathArr {
				vsetLogPath := fmt.Sprintf("%s-vset%d", logPath, vsetID)
				logDirArr = append(logDirArr, vsetLogPath)
			}
			for _, dplGroupID := range base.DPLGroupIDs {
				for _, logPath := range sv.LogPathArr {
					dplLogPath := fmt.Sprintf("%s-vset%d-%d", logPath, vsetID, dplGroupID)
					logDirArr = append(logDirArr, dplLogPath)
				}
			}
		}
	case Mjcachedpl.Type, Djcachedpl.Type: // *-vset1-1
		for _, vsetID := range base.VsetIDs {
			for _, logPath := range sv.LogPathArr {
				vsetLogPath := fmt.Sprintf("%s-vset%d", logPath, vsetID)
				logDirArr = append(logDirArr, vsetLogPath)
			}
			for _, jcacheGroupID := range base.JcacheGroupIDs {
				for _, logPath := range sv.LogPathArr {
					jcacheLogPath := fmt.Sprintf("%s-vset%d-%d", logPath, vsetID, jcacheGroupID)
					logDirArr = append(logDirArr, jcacheLogPath)
				}
			}
		}
	case Mcmapdpl.Type, Dcmapdpl.Type: // *-1
		for _, cmapGroupID := range base.CmapGroupIDs {
			for _, logPath := range sv.LogPathArr {
				vsetLogPath := fmt.Sprintf("%s-%d", logPath, cmapGroupID)
				logDirArr = append(logDirArr, vsetLogPath)
			}
		}
	case Flushdpl.Type, Vizions3.Type, Dpldagent.Type: // *-vset1
		for _, vsetID := range base.VsetIDs {
			for _, logPath := range sv.LogPathArr {
				vsetLogPath := fmt.Sprintf("%s-vset%d", logPath, vsetID)
				logDirArr = append(logDirArr, vsetLogPath)
			}
		}
	case Cdcgcbd.Type, Cdcgcs3.Type: // *-vset-1
		for _, vsetID := range base.VsetIDs {
			for _, logPath := range sv.LogPathArr {
				vsetLogPath := fmt.Sprintf("%s-vset-%d", logPath, vsetID)
				logDirArr = append(logDirArr, vsetLogPath)
			}
		}
	default: // by define in Service
		logDirArr = sv.LogPathArr
	}
	return
}

// ReverseServiceArr ...
func ReverseServiceArr(arr []Service) (reverseArr []Service) {
	length := len(arr)
	for i := 0; i < length; i++ {
		reverseArr = append(reverseArr, arr[length-1-i])
	}
	return
}

// ========== Define: Service/Binary ==========
// ========== DPL Service/Binary ==========
var (
	// Mjcachedpl .
	Mjcachedpl = Service{
		Name:       "mjcacheserver",
		Path:       "/opt/ccc/node/service/dpl/bin/mjcacheserver",
		GitPath:    "build/mjcacheserver",
		Type:       65537,
		TypeName:   "MJCACHE_SERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=mjcachedpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel:  "mjcachedpl=true", // k=v, v-<vset_id>-<group_id>
		Container:  "mjcachedpl",
		Replicas:   3,
		GetPid:     "ps -ax|grep -v grep|grep mjcacheserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-mjcachedpl"},
	}

	// Djcachedpl .
	Djcachedpl = Service{
		Name:       "djcacheserver",
		Path:       "/opt/ccc/node/service/dpl/bin/djcacheserver",
		GitPath:    "build/djcacheserver",
		Type:       8390609,
		TypeName:   "DJCACHE_SERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=djcachedpl-17", // k=v, v-<vset_id>-<group_id>
		NodeLabel:  "djcachedpl-17=true", // k=v, v-<vset_id>-<group_id>
		Container:  "djcachedpl",
		Replicas:   3,
		GetPid:     "ps -ax|grep -v grep|grep djcacheserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-djcachedpl-12", "pzcl-djcachedpl-17"},
	}

	// Jddpl .
	Jddpl = Service{
		Name:       "jdserver",
		Path:       "/opt/ccc/node/service/dpl/bin/jdserver",
		GitPath:    "build/jd",
		Type:       8388609,
		TypeName:   "JDSERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=jddpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel:  "jddpl=true", // k=v, v-<vset_id>-<group_id>
		Container:  "jddpl",
		Replicas:   3,
		GetPid:     "ps -ax|grep -v grep|grep jddpl|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-jddpl"},
	}

	// Servicedpl .
	Servicedpl = Service{
		Name:       "dplserver",
		Path:       "/opt/ccc/node/service/dpl/bin/dplserver",
		GitPath:    "build/dpl", // "build/server"
		Type:       1024,
		TypeName:   "DPLSERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=servicedpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel:  "servicedpl=true", // k=v, v-<vset_id>-<group_id>
		Container:  "servicedpl",
		Replicas:   3,
		GetPid:     "ps -ax|grep -v grep|grep servicedpl|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-servicedpl"},
	}

	// Flushdpl .
	Flushdpl = Service{
		Name:       "flushserver",
		Path:       "/opt/ccc/node/service/dpl/bin/flushserver",
		GitPath:    "build/flush",
		Type:       4194305,
		TypeName:   "FLUSHSERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=flushdpl", // k=v, v-<vset_id>
		NodeLabel:  "flushdpl=true", // k=v, v-<vset_id>
		Container:  "flushdpl",
		Replicas:   3,
		GetPid:     "ps -ax|grep -v grep|grep flushdpl|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-flushdpl"},
	}

	// Mcmapdpl .
	Mcmapdpl = Service{
		Name:       "mcmapserver",
		Path:       "/opt/ccc/node/service/dpl/bin/mcmapserver",
		GitPath:    "build/mcmap",
		Type:       2097153,
		TypeName:   "MCMAPSERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=cmapmcdpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel:  "cmapmcdpl=true", // k=v, v-<vset_id>-<group_id>
		Container:  "cmapmcdpl",
		Replicas:   0,
		GetPid:     "ps -ax|grep -v grep|grep mcmapserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-cmapdpl", "pzcl-cmapmcdpl"},
	}

	// Dcmapdpl .
	Dcmapdpl = Service{
		Name:       "dcmapserver",
		Path:       "/opt/ccc/node/service/dpl/bin/dcmapserver",
		GitPath:    "build/dcmap",
		Type:       8389609,
		TypeName:   "DCMAPSERVER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "name=cmapdcdpl-17", // k=v, v-<vset_id>-<group_id>
		NodeLabel:  "cmapdcdpl-17=true", // k=v, v-<vset_id>-<group_id>
		Container:  "cmapdcdpl",
		Replicas:   0,
		GetPid:     "ps -ax|grep -v grep|grep dcmapserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-cmapdcdpl-12", "pzcl-cmapdcdpl-17"},
	}

	// Dpldagent .
	Dpldagent = Service{
		Name:       "dpldagent",
		Path:       "/opt/ccc/node/service/dpl/bin/dpldagent",
		GitPath:    "build/dagent",
		Type:       524289,
		TypeName:   "DPLDAGENT",
		NameSpace:  "vizion",
		K8sKind:    K8sDaemonsets,
		PodLabel:   "name=bd-vset", // k=v, v-<vset_id>
		NodeLabel:  "bd-vset=true", // k=v, v-<vset_id>
		Container:  "bd",
		Replicas:   1,
		GetPid:     "ps -ax|grep -v grep|grep dpldagent|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-bd", "pzcl-bd-agent"},
	}

	// Vizions3 .
	Vizions3 = Service{
		Name:       "vizions3",
		Path:       "/opt/ccc/node/service/dpl/bin/vizions3",
		GitPath:    "src/s3/src/rgw",
		Type:       33,
		TypeName:   "S3",
		NameSpace:  "vizion",
		K8sKind:    K8sDeployment,
		PodLabel:   "name=vizion-s3-vset", // k=v, v-<vset_id>
		NodeLabel:  "vizion-s3-vset=true", // k=v, v-<vset_id>
		Container:  "vizions3",
		Replicas:   1,
		GetPid:     "ps -ax|grep -v grep|grep vizions3|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{"pzcl-s3"},
	}

	// Dplmanager .
	Dplmanager = Service{
		Name:       "dplmanager",
		Path:       "/opt/ccc/node/service/dpl/bin/dplmanager",
		GitPath:    "build/manager",
		Type:       34,
		TypeName:   "DPLMANAGER",
		NameSpace:  "vizion",
		K8sKind:    K8sDeployment,
		PodLabel:   "name=dplmanager",                   // k=v, v
		NodeLabel:  "node-role.kubernetes.io/node=true", // k=v, v
		Container:  "dplmanager",
		Replicas:   1,
		GetPid:     "ps -ax|grep -v grep|grep dplmanager|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{},
	}

	// Dplexporter .
	Dplexporter = Service{
		Name:       "dplexporter",
		Type:       35,
		TypeName:   "DPLEXPORTER",
		NameSpace:  "vizion",
		K8sKind:    K8sDeployment,
		PodLabel:   "name=dplexporter",                  // k=v, v
		NodeLabel:  "node-role.kubernetes.io/node=true", // k=v, v
		Container:  "dplexporter",
		Replicas:   1,
		GetPid:     "ps -ax|grep -v grep|grep dplexporter|grep -v bash|grep -v kubelet|awk '{print $1}'",
		LogPathArr: []string{},
	}
)

// ========== Define: DPL Binary ==========
var (
	// Dplko .
	Dplko = Service{
		Name:    "dpl.ko",
		Path:    "/opt/ccc/node/service/dpl/bin/dpl.ko",
		GitPath: "build/driver",
	}

	// Enctool .
	Enctool = Service{
		Name:    "enctool",
		Path:    "/opt/ccc/node/service/dpl/bin/enctool",
		GitPath: "build/enctool",
	}

	// Dplut .
	Dplut = Service{
		Name:    "dplut",
		Path:    "/opt/ccc/node/service/dpl/bin/dplut",
		GitPath: "build/ut",
	}

	// Libetcdv3 .
	Libetcdv3 = Service{
		Name:    "libetcdv3.so",
		Path:    "/usr/lib64/libetcdv3.so",
		GitPath: "build/libs/",
	}
)

// ========== Define: APP Service ==========
var (
	// ES .
	ES = Service{
		Name:       "es",
		Type:       2049,
		TypeName:   "ES",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "role=es-cold-data", // k=v, v-<vset_id>
		NodeLabel:  "elk=true",          // k=v, v-<vset_id>
		Container:  "es-cold-data",
		Replicas:   3,
		GetPid:     "ps -ax|grep -v grep|grep elasticsearch|grep java|awk '{print $1}'",
		LogPathArr: []string{},
	}

	// Nfsprovisioner .
	Nfsprovisioner = Service{
		Name:       "nfs",
		Type:       2050,
		TypeName:   "NFS_PROVISIONER",
		NameSpace:  "vizion",
		K8sKind:    K8sStatefulsets,
		PodLabel:   "app=nfs-provisioner",  // k=v, v-<vset_id>
		NodeLabel:  "nfs-provisioner=true", // k=v, v-<vset_id>
		Container:  "nfs-provisioner",
		Replicas:   1,
		GetPid:     "ps -ax|grep -v grep|grep nfs-provisioner|grep java|awk '{print $1}'",
		LogPathArr: []string{},
	}

	// Cdcgcs3 .
	Cdcgcs3 = Service{
		Name:       "cdcgcs3",
		Type:       2051,
		TypeName:   "CDCGC_S3",
		NameSpace:  "vizion",
		K8sKind:    K8sDeployment,
		PodLabel:   "run=s3-cdcgc-vset",                 // k=v, v-<vset_id>
		NodeLabel:  "node-role.kubernetes.io/node=true", // k=v, v-<vset_id>
		Container:  "cdcgc",
		Replicas:   1,
		GetPid:     "",
		LogPathArr: []string{"pzcl-s3-cdcgc-log", "pzcl-s3-cdcgc-data"},
	}

	// Cdcgcbd .
	Cdcgcbd = Service{
		Name:       "cdcgcbd",
		Type:       2052,
		TypeName:   "CDCGC_BD",
		NameSpace:  "vizion",
		K8sKind:    K8sDeployment,
		PodLabel:   "run=bd-cdcgc-vset",                 // k=v, v-<vset_id>
		NodeLabel:  "node-role.kubernetes.io/node=true", // k=v, v-<vset_id>
		Container:  "cdcgc",
		Replicas:   1,
		GetPid:     "",
		LogPathArr: []string{"pzcl-bd-cdcgc-log", "pzcl-bd-cdcgc-data"},
	}
)

// ========== Define: MASTER Services ==========
var (
	// ETCD .
	ETCD = Service{
		Name:      "ETCD",
		Type:      101,
		TypeName:  "ETCD",
		NameSpace: "kube-system",
		K8sKind:   "",
		PodLabel:  "component=etcd",
		NodeLabel: "node-role.kubernetes.io/etcd=true",
		Container: "",
		Replicas:  3,
		GetPid:    "",
	}

	// MasterCass .
	MasterCass = Service{
		Name:      "MasterCass",
		Type:      102,
		TypeName:  "MasterCass",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "app=cassandra-master", // k=v, v
		NodeLabel: "cassandra-0=true",     // k=v, v
		Container: "cassandra",
		Replicas:  3,
		GetPid:    "",
	}

	// SubCass .
	SubCass = Service{
		Name:      "SubCass",
		Type:      103,
		TypeName:  "SubCass",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "app=cassandra-vset", // k=v, v-<vset_id>
		NodeLabel: "cassandra=true",     // k=v, v-<vset_id>
		Container: "cassandra",
		Replicas:  3,
		GetPid:    "",
	}

	// CassMonitor .
	CassMonitor = Service{
		Name:      "CassandraMonitor",
		Type:      104,
		TypeName:  "CassandraMonitor",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "app=cassandra-monitor",  // k=v
		NodeLabel: "cassandra-monitor=true", // k=v
		Container: "cassandra",
		Replicas:  1,
		GetPid:    "",
	}

	// MysqlCluster .
	MysqlCluster = Service{
		Name:      "MysqlCluster",
		Type:      105,
		TypeName:  "MysqlCluster",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "v1alpha1.mysql.oracle.com/cluster=mysql-cluster", // k=v, v
		NodeLabel: "label-mysql=true",                                // k=v, v
		Container: "mysql",
		Replicas:  3,
		GetPid:    "",
	}

	// MysqlOperator .
	MysqlOperator = Service{
		Name:      "MysqlOperator",
		Type:      106,
		TypeName:  "MysqlOperator",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "app=mysql-operator", // k=v, v
		NodeLabel: "label-mysql=true",   // k=v, v
		Container: "mysql",
		Replicas:  1,
		GetPid:    "",
	}

	// MysqlRouter .
	MysqlRouter = Service{
		Name:      "MysqlRouter",
		Type:      107,
		TypeName:  "MysqlRouter",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "app=mysql-router", // k=v, v
		NodeLabel: "label-mysql=true", // k=v, v
		Container: "mysql",
		Replicas:  2,
		GetPid:    "",
	}
)

// ========== Define: Clean Item ==========
var (
	CleanLog = CleanItem{
		Name: "log",
		Arg:  nil,
	}

	CleanJdevice = CleanItem{
		Name: "j_device",
		Arg:  nil,
	}

	CleanSC = CleanItem{
		Name: "storage_cache",
		Arg: []string{
			"/opt/storage_cache/",
		},
	}

	CleanCdcgc = CleanItem{
		Name: "cdcgc",
		Arg:  nil,
	}

	CleanMasterCass = CleanItem{
		Name: "master_cassandra",
		Arg:  nil,
	}

	CleanSubCass = CleanItem{
		Name: "sub_cassandra",
		Arg: []string{
			"vizion_ns_vol.vbns",
			"vizion_ns_vol.vinode",
			"vizion_ns_s3.inode",
			"vizion_ns_s3.s3part",
			"vizion_ns_s3.s3ns",
			"vizion_ns_common.fp",
			"vizion_ns_common.fp_deletion",
		},
	}

	CleanEtcd = CleanItem{
		Name: "etcd",
		Arg: []string{
			"/vizion/dpl/add_vol",
			"/vizion/dpl/stg_unit",
			"/vizion/dpl/jnl",
		},
	}
)

// ========== Default: Service/Binary ==========
var (
	// Default DPL Service for test/upgrade  -- order by start
	DefaultDplServiceArray = []Service{
		Dplmanager,
		Jddpl,
		Servicedpl,
		Mjcachedpl,
		Djcachedpl,
		Flushdpl,
		Mcmapdpl,
		Dcmapdpl,
		Vizions3,
		Dpldagent,
		Dplexporter,
	}

	// Default DPL Binary for upgrade  -- order by start
	DefaultDplBinaryArray = []Service{
		Dplmanager,
		Jddpl,
		Servicedpl,
		Mjcachedpl,
		Djcachedpl,
		Flushdpl,
		Mcmapdpl,
		Dcmapdpl,
		Vizions3,
		Dpldagent,
		Dplexporter,

		Dplko,
		Enctool,
		Dplut,
		Libetcdv3,
	}

	// Default APP Service for test  -- order by start
	DefaultAppServiceArray = []Service{
		ES,
		Nfsprovisioner,
		Cdcgcs3,
		Cdcgcbd,
	}

	// Default MASTER service for test
	DefaultMasterServiceArray = []Service{
		ETCD,
		MasterCass,
		SubCass,
		MysqlCluster,
		MysqlOperator,
		MysqlRouter,
	}

	// Default Core Service Array: DPL + APP  -- order by start
	DefaultCoreServiceArray = append(DefaultDplServiceArray, DefaultAppServiceArray...)

	// Default Service Array: DPL + APP + MASTER
	DefaultServiceArray = append(DefaultMasterServiceArray, DefaultCoreServiceArray...)
)

// ========== Default: CleanItem ==========
var (
	// DefaultCleanArray define the default cleanup item for maint/upgrade
	DefaultCleanArray = []CleanItem{
		CleanLog,
		CleanJdevice,
		CleanEtcd,
		CleanSC,
		// CleanMasterCass,  // Not need clean master cass now
		CleanSubCass,
		CleanCdcgc,
	}
)

// DefaultCHTYPEArray define the default CH_TYPE list
var DefaultCHTYPEArray = []string{CHTYPEBD, CHTYPES3}
