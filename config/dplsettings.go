package config

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

// DPL/SERVICE Settings
const (
	CHTYPEBD = "CH_TYPE_BD"
	CHTYPES3 = "CH_TYPE_S3"
)

// Channel define the channel on dpl
type Channel struct {
	Type string
	ID   string
}

// Service define the service information
type Service struct {
	Name      string // service name
	Path      string // service binary path in pod
	GitPath   string // servicepath in gitlab
	Type      int    // service Type ID
	TypeName  string // service type name
	NameSpace string // service in k8s namespace
	K8sKind   string // service resource kind in k8s, sts,ds,deploy...
	PodLabel  string // pod label template
	NodeLabel string // node label template
	Container string // service pod container name
	Replicas  int    // service replicas
	GetPid    string // the cmd to get service pid, eg: ps -ax | grep dpl
}

// CleanItem define the cleanup item
type CleanItem struct {
	Name string
	Arg  []string
}

// ========== DPL Service/Binary ==========
var (
	// Mjcachedpl .
	Mjcachedpl = Service{
		Name:      "mjcacheserver",
		Path:      "/opt/ccc/node/service/dpl/bin/mjcacheserver",
		GitPath:   "build/mjcacheserver",
		Type:      65537,
		TypeName:  "MJCACHE_SERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=mjcachedpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel: "mjcachedpl=true", // k=v, v-<vset_id>-<group_id>
		Container: "mjcachedpl",
		Replicas:  3,
		GetPid:    "ps -ax|grep -v grep|grep mjcacheserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Djcachedpl .
	Djcachedpl = Service{
		Name:      "djcacheserver",
		Path:      "/opt/ccc/node/service/dpl/bin/djcacheserver",
		GitPath:   "build/djcacheserver",
		Type:      8390609,
		TypeName:  "DJCACHE_SERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=djcachedpl-17", // k=v, v-<vset_id>-<group_id>
		NodeLabel: "djcachedpl-17=true", // k=v, v-<vset_id>-<group_id>
		Container: "djcachedpl",
		Replicas:  3,
		GetPid:    "ps -ax|grep -v grep|grep djcacheserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Jddpl .
	Jddpl = Service{
		Name:      "jdserver",
		Path:      "/opt/ccc/node/service/dpl/bin/jdserver",
		GitPath:   "build/jd",
		Type:      8388609,
		TypeName:  "JDSERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=jddpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel: "jddpl=true", // k=v, v-<vset_id>-<group_id>
		Container: "jddpl",
		Replicas:  3,
		GetPid:    "ps -ax|grep -v grep|grep jddpl|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Servicedpl .
	Servicedpl = Service{
		Name:      "dplserver",
		Path:      "/opt/ccc/node/service/dpl/bin/dplserver",
		GitPath:   "build/dpl", // "build/server"
		Type:      1024,
		TypeName:  "DPLSERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=servicedpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel: "servicedpl=true", // k=v, v-<vset_id>-<group_id>
		Container: "servicedpl",
		Replicas:  3,
		GetPid:    "ps -ax|grep -v grep|grep servicedpl|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Flushdpl .
	Flushdpl = Service{
		Name:      "flushserver",
		Path:      "/opt/ccc/node/service/dpl/bin/flushserver",
		GitPath:   "build/flush",
		Type:      4194305,
		TypeName:  "FLUSHSERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=flushdpl", // k=v, v-<vset_id>
		NodeLabel: "flushdpl=true", // k=v, v-<vset_id>
		Container: "flushdpl",
		Replicas:  3,
		GetPid:    "ps -ax|grep -v grep|grep flushdpl|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Mcmapdpl .
	Mcmapdpl = Service{
		Name:      "mcmapserver",
		Path:      "/opt/ccc/node/service/dpl/bin/mcmapserver",
		GitPath:   "build/mcmap",
		Type:      2097153,
		TypeName:  "MCMAPSERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=cmapmcdpl", // k=v, v-<vset_id>-<group_id>
		NodeLabel: "cmapmcdpl=true", // k=v, v-<vset_id>-<group_id>
		Container: "cmapmcdpl",
		Replicas:  0,
		GetPid:    "ps -ax|grep -v grep|grep mcmapserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Dcmapdpl .
	Dcmapdpl = Service{
		Name:      "dcmapserver",
		Path:      "/opt/ccc/node/service/dpl/bin/dcmapserver",
		GitPath:   "build/dcmap",
		Type:      8389609,
		TypeName:  "DCMAPSERVER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "name=cmapdcdpl-17", // k=v, v-<vset_id>-<group_id>
		NodeLabel: "cmapdcdpl-17=true", // k=v, v-<vset_id>-<group_id>
		Container: "cmapdcdpl",
		Replicas:  0,
		GetPid:    "ps -ax|grep -v grep|grep dcmapserver|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Dpldagent .
	Dpldagent = Service{
		Name:      "dpldagent",
		Path:      "/opt/ccc/node/service/dpl/bin/dpldagent",
		GitPath:   "build/dagent",
		Type:      524289,
		TypeName:  "BLOCK_DEVICE",
		NameSpace: "vizion",
		K8sKind:   K8sDaemonsets,
		PodLabel:  "name=bd-vset", // k=v, v-<vset_id>
		NodeLabel: "bd-vset=true", // k=v, v-<vset_id>
		Container: "bd",
		Replicas:  1,
		GetPid:    "ps -ax|grep -v grep|grep dpldagent|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Vizions3 .
	Vizions3 = Service{
		Name:      "vizions3",
		Path:      "/opt/ccc/node/service/dpl/bin/vizions3",
		GitPath:   "src/s3/src/rgw",
		Type:      8389609,
		TypeName:  "S3",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "name=vizion-s3-vset", // k=v, v-<vset_id>
		NodeLabel: "vizion-s3-vset=true", // k=v, v-<vset_id>
		Container: "vizions3",
		Replicas:  1,
		GetPid:    "ps -ax|grep -v grep|grep vizions3|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Dplmanager .
	Dplmanager = Service{
		Name:      "dplmanager",
		Path:      "/opt/ccc/node/service/dpl/bin/dplmanager",
		GitPath:   "build/manager",
		Type:      34,
		TypeName:  "S3",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "name=dplmanager",                   // k=v, v
		NodeLabel: "node-role.kubernetes.io/node=true", // k=v, v
		Container: "dplmanager",
		Replicas:  1,
		GetPid:    "ps -ax|grep -v grep|grep dplmanager|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}

	// Dplexporter .
	Dplexporter = Service{
		Name:      "dplexporter",
		Type:      35,
		TypeName:  "DPLEXPORTER",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "name=dplexporter",                  // k=v, v
		NodeLabel: "node-role.kubernetes.io/node=true", // k=v, v
		Container: "dplexporter",
		Replicas:  1,
		GetPid:    "ps -ax|grep -v grep|grep dplexporter|grep -v bash|grep -v kubelet|awk '{print $1}'",
	}
)

// ========== DPL Binary ==========
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

// ========== APP Service ==========
var (
	// ES .
	ES = Service{
		Name:      "es",
		Type:      2049,
		TypeName:  "ES",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "role=es-cold-data", // k=v, v-<vset_id>
		NodeLabel: "elk=true",          // k=v, v-<vset_id>
		Container: "es-cold-data",
		Replicas:  3,
		GetPid:    "ps -ax|grep -v grep|grep elasticsearch|grep java|awk '{print $1}'",
	}

	// Nfsprovisioner .
	Nfsprovisioner = Service{
		Name:      "nfs-provisioner",
		Type:      2050,
		TypeName:  "NFS_PROVISIONER",
		NameSpace: "vizion",
		K8sKind:   K8sStatefulsets,
		PodLabel:  "app=nfs-provisioner",  // k=v, v-<vset_id>
		NodeLabel: "nfs-provisioner=true", // k=v, v-<vset_id>
		Container: "nfs-provisioner",
		Replicas:  1,
		GetPid:    "ps -ax|grep -v grep|grep nfs-provisioner|grep java|awk '{print $1}'",
	}

	// Cdcgcs3 .
	Cdcgcs3 = Service{
		Name:      "cdcgc-s3",
		Type:      2051,
		TypeName:  "CDCGC_S3",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "run=s3-cdcgc-vset",                 // k=v, v-<vset_id>
		NodeLabel: "node-role.kubernetes.io/node=true", // k=v, v-<vset_id>
		Container: "cdcgc",
		Replicas:  1,
		GetPid:    "",
	}

	// Cdcgcbd .
	Cdcgcbd = Service{
		Name:      "cdcgc-bd",
		Type:      2052,
		TypeName:  "CDCGC_BD",
		NameSpace: "vizion",
		K8sKind:   K8sDeployment,
		PodLabel:  "run=bd-cdcgc-vset",                 // k=v, v-<vset_id>
		NodeLabel: "node-role.kubernetes.io/node=true", // k=v, v-<vset_id>
		Container: "cdcgc",
		Replicas:  1,
		GetPid:    "",
	}
)

// ========== MASTER Services ==========
var (
	// ETCD .
	ETCD = Service{
		Name:      "ETCD",
		Type:      101,
		TypeName:  "ETCD",
		NameSpace: "kube-system",
		K8sKind:   "",
		PodLabel:  "component=etcd", // k=v, v-<vset_id>
		NodeLabel: "node-role.kubernetes.io/etcd=true",          // k=v, v-<vset_id>
	}
)


// ========== Clean Item ==========
var (
	CleanLog = CleanItem{
		Name: "log",
		Arg:  nil,
	}

	CleanJournal = CleanItem{
		Name: "journal",
		Arg:  nil,
	}

	CleanSC = CleanItem{
		Name: "storage_cache",
		Arg:  []string{
			"/opt/storage_cache/"
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
		Arg:  []string{
			"/vizion/dpl/add_vol"
		},
	}
)

// DefaultCHTYPEArray define the default CH_TYPE list
var DefaultCHTYPEArray = []string{CHTYPEBD, CHTYPES3}
