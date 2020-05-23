package config

// K8S Settings
const (
	K8sDeployment   = "deployment"
	K8sStatefulsets = "statefulsets"
	K8sDaemonsets   = "daemonsets"
)

// Upgrade Image Settings
const (
	DplBuildIP        = "10.199.116.1"
	DplBuildPath      = "/home/project/dpl/develop/dpl"
	DplBuildLocalPath = "/mnt/vizion/QA/dpl/"
	RemoteRegistryIP  = "10.180.1.45"
	RemoteRegistry    = "registry.vizion.ai"
	RemoteDplRegistry = "registry.vizion.ai/stable/dpl"
	LocalRegistry     = "registry.vizion.local"
	LocalDplRegistry  = "registry.vizion.local/stable/dpl"
	DefaultDplImage   = "registry.vizion.local/library/dpl:tmp"
)

// DPL/SERVICE Settings
const (
	DplmanagerLocalPath = "/usr/bin/dplmanager"
	JDevicePath         = "/dev/j_device"
	CHTYPEBD            = "CH_TYPE_BD"
	CHTYPES3            = "CH_TYPE_S3"
)

// Service define the dpl service information
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

// Define the DPL Service informations
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
		Type:      201,
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
		Type:      202,
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

// Define the APP Service informations
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
)

// DefaultCHTYPEArray define the default CH_TYPE list
var DefaultCHTYPEArray = []string{CHTYPEBD, CHTYPES3}
