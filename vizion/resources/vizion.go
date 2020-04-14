package resources

// CassandraRes ...
type CassandraRes interface {
}

// ServiceRes ...
type ServiceRes interface {
	DiffPodsStatus(lastStatusArray []string, diffNode, diffRestart bool, ignorePods, ignoreNodes []string, serviceTypeArray []int)
}
