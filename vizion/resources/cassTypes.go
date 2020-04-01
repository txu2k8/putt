package resources

import (
	"time"
)

// CassandraCluster ...
type CassandraCluster struct {
	Idx             int               `db:"idx"`
	CaKeySpace      []int             `db:"ca_keyspace"`
	ContactIps      []string          `db:"contact_ips"`
	DaemonSetID     string            `db:"daemonset_id"`
	Data            map[string]string `db:"data"`
	DataStorageType int               `db:"data_storage_type"`
	Info            string            `db:"info"`
	Lock            int               `db:"lock"`
	Password        string            `db:"password"`
	Service         []string          `db:"service"`
	Status          int               `db:"status"`
	Type            int               `db:"type"`
	User            string            `db:"user"`
}

// StorageRow ...
type StorageRow struct {
	Index               int  `db:"idx"`
	DefaultStorageIndex bool `db:"default_storage_index"`
}

// MakeNodeInput ...
type MakeNodeInput struct {
	DeviceName      string
	VolumeID        string
	VolumeName      string
	VolumeIdx       int64
	BdServiceIP     string
	BdServicePort   int
	MakeVolumeSize  int
	VsetID          int
	NeedReCreate    bool
	ServiceDPl      *Service
	VolumeParameter MountParameter
}

const (
	// DPLHealthCheck ...
	DPLHealthCheck string = "dpl -a %s -p %d helo"

	// DPLFlush ...
	DPLFlush string = "vpm -a %s -p %d jfs flush"

	// DPLCheckFlushStatus ...
	DPLCheckFlushStatus string = "vpm -a %s -p %d jfs flush stat %s"

	// ResizeCommand ...
	ResizeCommand string = "bd -a %s -p %d -d %s -s %d resize"
	// AddDeviceCmd ...
	AddDeviceCmd string = "%s -m bd -a %s -p %d -s %d -u %s -v %s --vset %d %s -d %s add"
	// AddDeviceCmdWithDPl ...
	AddDeviceCmdWithDPl string = "%s -m bd -a %s -p %d -s %d -u %s -v %s --preferred_dpl %s:%d --vset %d -d /dev/%s add"
	// DeleteDeviceCmd ...
	DeleteDeviceCmd string = "bd -a %s -p %d -d %s delete"
	// DplManagerPath ...
	DplManagerPath string = "dplmanager"
	// BDSERVICE ...
	BDSERVICE int = 524289
	// DPLSERVICE ...
	DPLSERVICE int = 1024
	// IndexedClonedStatus ...
	IndexedClonedStatus int = 3
	// VPMDPLSERVICE ...
	VPMDPLSERVICE int = 131073

	// EnableMarkerCMD ...
	EnableMarkerCMD string = "ch -a %s -p %s -c %s pxs marker enable"
	// DisableMarkerCMD ...
	DisableMarkerCMD string = "ch -a %s -p %s -c %s pxs marker disable"
	// StatMarkerCMD ...
	StatMarkerCMD string = "ch -a %s -p %s -c %s pxs marker stat"
	// ResizeDeviceCmd ...
	ResizeDeviceCmd string = "bd -a %s -p%d -d %s -s %s resize"

	// DeleteBucketFromCache ...
	DeleteBucketFromCache string = "s3 -a %s -p %d cc_bucket '%s'"

	// DeleteVolumeFromCache ...
	DeleteVolumeFromCache string = "s3 -a %s -p %d cc_volume '%s'"

	// EntityIndexMapID ...
	EntityIndexMapID string = "00000000-0000-0000-4444-111111111111"

	// DefaultMaxVolumeSnapshotNum ...
	DefaultMaxVolumeSnapshotNum int = 3
)
const (
	//MgrGeneralError ...
	MgrGeneralError int = 100
	//MgrAexNotWorking ..
	MgrAexNotWorking = 99
	//MgrNoDevice ...
	MgrNoDevice = 98
	//MgrDeviceNotStarted ...
	MgrDeviceNotStarted = 97
	//MgrDeviceAlreadyAdded ...
	MgrDeviceAlreadyAdded = 96
	//MgrFailedConnect ...
	MgrFailedConnect = 95
	//MgrFailedSendReq ...
	MgrFailedSendReq = 94
	//MgrNotDagent ...
	MgrNotDagent = 93
	//MgrNoResp ...
	MgrNoResp = 92
	//MgrWrongDevPath ...
	MgrWrongDevPath = 91
	//MgrInitError ...
	MgrInitError = 255
	//MgrDeviceKeyExist ...
	MgrDeviceKeyExist = 90
)

// DPLTypeMap ...
var DPLTypeMap = map[string]int64{
	"s3":          33,
	"service_dpl": 1024,
	"vpm_dpl":     131073,
	"anchor_dpl":  65537,
	"bd_dpl":      524289,
	"cmap_dpl":    2097153,
	"flush_dpl":   4194305,
	"jd_dpl":      8388609,
}

// SnapshotStatus ...
var SnapshotStatus = map[string]int{
	"start":                0,
	"creating":             1,
	"unfreeze":             2,
	"record":               3,
	"accomplished":         4,
	"statFailed":           -1,
	"createSnapshotFailed": -2,
	"disableMarketFailed":  -3,
	"forceDelete":          -4,
	"thawFailed":           -5,
}
var (
	// ReturnCodeForCreatingBucket ....
	ReturnCodeForCreatingBucket = map[string]int{
		"SUCCESS":                   0,
		"FAIL":                      -1,
		"LOCK_NAME_FAIL":            1,
		"NOT_SUPPORTED_NAME":        2,
		"INCORRECT_USER":            3,
		"USER_MAX_BUCKET_NUM_REACH": 4,
	}
	// ReturnCodeForDeletingBucket ....
	ReturnCodeForDeletingBucket = map[string]int{
		"SUCCESS":        0,
		"FAIL":           -1,
		"NO_SUCH_BUCKET": 1,
	}
	// CassandraTypeClusterMap ...
	CassandraTypeClusterMap = map[int]string{
		0: "MASTER",
		1: "KAIROSDB",
		2: "NS_S3",
		3: "NS_CC",
		4: "NS_VB",
	}

	// VolumeFormatMap ...
	VolumeFormatMap = map[string]string{
		"ext4": "mkfs.ext4",
		"ext3": "mkfs.ext3",
	}

	// EntityTypeMap ...
	EntityTypeMap = map[string]int{
		"BUCKET": 1,
		"VOLUME": 2,
	}

	// BucketTypeMap ...
	BucketTypeMap = map[string]int{
		"REGULAR":  0,
		"SNAPSHOT": 1,
		"ROLLBACK": 2,
	}

	// VolumeTypeMap ...
	VolumeTypeMap = map[string]int{
		"REGULAR":  0,
		"SNAPSHOT": 1,
		"ROLLBACK": 2,
	}
)

// Policy ...
type Policy struct {
	Type           int               `db:"type"`
	Tenant         string            `db:"tenant"`
	Name           string            `db:"name"`
	Action         string            `db:"action"`
	AttachedEntity []string          `db:"attached_entity"`
	Data           map[string]string `db:"data"`
	Readonly       int               `db:"readonly"`
	Schedule       string            `db:"schedule"`
	ScheduleType   int               `db:"schedule_type"`
}

// PolicyAssociation ...
type PolicyAssociation struct {
	EntityName       string `db:"entity_name"`
	EntityType       int    `db:"entity_type"`
	EntityParentName string `db:"entity_parent_name"`
	Tenant           string `db:"tenant"`
	PolicyType       int    `db:"policy_type"`
	PolicyName       string `db:"policy_name"`
	Info             string `db:"info"`
}

// S3User ...
type S3User struct {
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

// Service ...
type Service struct {
	Type        int               `db:"type"`
	ID          string            `db:"id"`
	CaCluster   string            `db:"ca_cluster"`
	Container   string            `db:"container"`
	Data        map[string]string `db:"data"`
	DataSet     []string          `db:"data_set"`
	EnableHa    bool              `db:"enable_ha"`
	IP          string            `db:"ip"`
	Location    string            `db:"location"`
	Name        string            `db:"name"`
	Node        string            `db:"node"`
	Pod         string            `db:"pod"`
	Port        int               `db:"port"`
	ServicePair map[string]int    `db:"service_pair"`
	Status      int               `db:"status"`
	Utime       time.Time         `db:"u_time"`
	Version     string            `db:"version"`
	VsetID      int               `db:"vset_id"`
}

// Tenant ...
type Tenant struct {
	Name   string      `db:"name"`
	MdList map[int]int `db:"md_list"`
}

// StorageProvider ...
type StorageProvider struct {
	VendorType     int            `db:"vendor_type"`
	ID             string         `db:"id"`
	ConnectionType int            `db:"connection_type"`
	Host           string         `db:"host"`
	Info           string         `db:"info"`
	Name           string         `db:"name"`
	Password       string         `db:"password"`
	Port           int            `db:"port"`
	Status         int            `db:"status"`
	StorageTarget  map[int]string `db:"storage_target"`
	StorageType    int            `db:"storage_type"`
	User           string         `db:"user"`
}

// StorageRegion ...
type StorageRegion struct {
	VendorType int    `db:"vendor_type"`
	RegionType int    `db:"region_type"`
	Endpoint   string `db:"endpoint"`
	Info       string `db:"info"`
	Protocol   string `db:"protocol"`
	Region     string `db:"region"`
	RegionName string `db:"region_name"`
	Status     int    `db:"status"`
}

// Storage ...
type Storage struct {
	Index               int               `db:"idx"`
	Data                map[string]string `db:"data"`
	DefaultStorageIndex bool              `db:"default_storage_index"`
	EncryptionType      int               `db:"encryption_type"`
	Info                string            `db:"info"`
	Name                string            `db:"name"`
	Path                string            `db:"path"`
	Region              int               `db:"region"`
	Status              int               `db:"status"`
	StorageProviderID   string            `db:"storage_provider_id"`
	Target              string            `db:"target"`
	Tenant              string            `db:"tenant"`
	VendorType          int               `db:"vendor_type"`
	BsShift             uint8             `db:"bs_shift"`
	Version             int               `db:"version"`
}

// StorageDc ...
type StorageDc struct {
	DcName              string            `db:"dc_name"`
	Index               int               `db:"idx"`
	Data                map[string]string `db:"data"`
	DefaultStorageIndex bool              `db:"default_storage_index"`
	EncryptionType      int               `db:"encryption_type"`
	Info                string            `db:"info"`
	Name                string            `db:"name"`
	Path                string            `db:"path"`
	Region              int               `db:"region"`
	Status              int               `db:"status"`
	StorageProviderID   string            `db:"storage_provider_id"`
	Target              string            `db:"target"`
	Tenant              string            `db:"tenant"`
	VendorType          int               `db:"vendor_type"`
	BsShift             uint8             `db:"bs_shift"`
	Version             int               `db:"version"`
}

// S3Ns ...
type S3Ns struct {
	BucketName string         `db:"bucket"`
	Path       string         `db:"path"`
	Offset     int            `db:"offset"`
	File       string         `db:"file"`
	UTime      time.Time      `db:"u_time"`
	Etag       string         `db:"etag"`
	Fp         string         `db:"fp"`
	Grantee    map[string]int `db:"grantee"`
	Inode      int            `db:"inode"`
	Owner      string         `db:"owner"`
	Size       int            `db:"size"`
	Status     int            `db:"status"`
}

// S3BucketGroup ...
type S3BucketGroup struct {
	Name    string    `db:"name"`
	Type    int       `db:"type"`
	CTime   time.Time `db:"c_time"`
	Buckets []string  `db:"bucket"`
	Owner   string    `db:"owner"`
	Status  int       `db:"status"`
}

// Node ...
type Node struct {
	Name      string   `db:"name"`
	ID        string   `db:"id"`
	Type      int      `db:"type"`
	StorageIP []string `db:"storage_ip"`
	HostIP    []string `db:"host_ip"`
}

// Volume ...
type Volume struct {
	Name               string         `db:"name"`
	Type               int            `db:"type"`
	BaseName           string         `db:"basename"`
	BsShift            uint8          `db:"bs_shift"`
	FileSystemName     string         `db:"file_system_name"`
	GcID               int            `db:"gc_id"`
	GcStatus           int            `db:"gc_status"`
	GcThreadName       string         `db:"gc_thread_name"`
	GcUTime            time.Time      `db:"gc_u_time"`
	Grantee            map[string]int `db:"grantee"`
	MdIndex            int            `db:"md_index"`
	MdType             int            `db:"md_type"`
	NodeID             string         `db:"node_id"`
	Owner              string         `db:"owner"`
	PairStorage        int            `db:"pair_storage"`
	PairVolume         string         `db:"pair_volume"`
	Property           int            `db:"property"`
	SnapName           string         `db:"snapname"`
	SubVolume          []string       `db:"sub_volume"`
	SuperName          string         `db:"supername"`
	UTime              time.Time      `db:"u_time"`
	VinodeLen          int            `db:"vinode_len"`
	VolumeGroup        []string       `db:"volumegroup"`
	Status             int            `db:"status"`
	Size               int            `db:"size"`
	BlockDeviceName    string         `db:"block_device_name"`
	BlockDeviceService string         `db:"block_device_service"`
	JournalDeviceName  string         `db:"journal_device_name"`
	PartitionJournal   bool           `db:"partition_journal"`
	Ctime              time.Time      `db:"c_time"`
	ID                 string         `db:"id"`
	Idx                int64          `db:"idx"`
	Index              []int          `db:"storage_index"`
	EnableJournal      bool           `db:"partition_journal"`
	Format             bool           `db:"format"`
	FormatType         string         `db:"format_type"`
	TbhLow             time.Time      `db:"tbh_low"`
	TbhHigh            time.Time      `db:"tbh_high"`
	CsiFlag            string         `db:"csi_flag"`
}

// SnapshotVolume ...
type SnapshotVolume struct {
	Volume
}

// BdConfig ...
type BdConfig struct {
	GlobalConf Global      `json:"global"`
	EtcdConf   interface{} `json:"etcd"`
	VpmConf    interface{} `json:"vpm"`
}

// Global ...
type Global struct {
	ServiceType int    `json:"service_type"`
	Vset        int    `json:"vset"`
	Type        string `json:"type"`
	UUID        string `json:"uuid"`
}

// BucketResponse ...
type BucketResponse struct {
	Code    int          `json:"code"`
	Data    OutPutBucket `json:"data"`
	Message string       `json:"message"`
}

// OutPutBucket ...
type OutPutBucket struct {
	Stderr string `json:"stderr"`
	Stdout string `json:"stdout"`
}

// APIResponse ...
type APIResponse struct {
	Code    int        `json:"code"`
	Data    OutPutData `json:"data"`
	Message string     `json:"message"`
}

// OutPutData ...
type OutPutData struct {
	Stderr string `json:"stderr"`
	Stdout string `json:"stdout"`
}

// IndexMap ...
type IndexMap struct {
	ID  string `db:"id"`
	Idx int64  `db:"idx"`
}

// IdxName ...
type IdxName struct {
	Idx  int64  `db:"idx"`
	Name string `db:"name"`
}

// Dc ...
type Dc struct {
	Name                  string `db:"name"`
	BucketVolumeIndexID   string `db:"bucket_volume_index_id"`
	BucketVolumeIndexHigh int64  `db:"bucket_volume_index_high"`
	BucketVolumeIndexLow  int64  `db:"bucket_volume_index_low"`
}

// MountParameter ...
type MountParameter struct {
	Policies VolumePolicies `json:"policies" from:"policies"`
}

// VolumePolicies ...
type VolumePolicies []map[string]interface{}

// ChannelInfo ...
type ChannelInfo struct {
	DeviceName         string `json:"dpl_device_name"`
	DplIP              string `json:"dpl_ip"`
	DplPort            string `json:"dpl_port"`
	DplID              string `json:"channel_uuid"`
	LastConnectionTime string `json:"last_connection_time"`
	ConnectionCounter  string `json:"connection_counter"`
}

// Bucket ...
type Bucket struct {
	BucketName     string            `db:"name"`
	Type           int               `db:"type"`
	CTime          time.Time         `db:"c_time"`
	Utime          time.Time         `db:"u_time"`
	BaseName       string            `db:"basename"`
	BucketGroup    []string          `db:"bucketgroup"`
	EncryptionType int               `db:"encryption_type"`
	GcID           int               `db:"gc_id"`
	GcStatus       int               `db:"gc_status"`
	GcThreadName   string            `db:"gc_thread_name"`
	GcUTime        time.Time         `db:"gc_u_time"`
	Grantee        map[string]int    `db:"grantee"`
	Idx            int64             `db:"idx"`
	MdIndex        int               `db:"md_index"`
	Property       int               `db:"property"`
	OwnerName      string            `db:"owner"`
	PairBucket     string            `db:"pair_bucket"`
	Status         int               `db:"status"`
	SuperName      string            `db:"supername"`
	StorageIndex   []int             `db:"storage_index"`
	SubBucket      []string          `db:"sub_bucket"`
	Usage          int               `db:"usage"`
	PairStorage    int               `db:"pair_storage"`
	Data           map[string]string `db:"data"`
	SnapshotName   string            `db:"snapname"`
	BsShift        uint8             `db:"bs_shift"`
}

// SnapshotBucket ...
type SnapshotBucket struct {
	Bucket
}

// RollBucket ...
type RollBucket struct {
	Bucket
}
