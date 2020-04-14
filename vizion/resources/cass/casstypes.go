package cass

import (
	"time"
)

// ============== master cassandra tables ==============

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

// Node ...
type Node struct {
	Name      string   `db:"name"`
	ID        string   `db:"id"`
	Type      int      `db:"type"`
	StorageIP []string `db:"storage_ip"`
	HostIP    []string `db:"host_ip"`
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

// Tenant ...
type Tenant struct {
	Name   string      `db:"name"`
	MdList map[int]int `db:"md_list"`
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

// S3Bucket ...
type S3Bucket struct {
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

// S3BucketGroup ...
type S3BucketGroup struct {
	Name    string    `db:"name"`
	Type    int       `db:"type"`
	CTime   time.Time `db:"c_time"`
	Buckets []string  `db:"bucket"`
	Owner   string    `db:"owner"`
	Status  int       `db:"status"`
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

// ============== vset sub cassandra tables ==============

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
