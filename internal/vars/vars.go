package vars

type DiskMetrics struct {
	PVName            string
	PVTotalSize       int64
	PVUsedSize        int64
	PVUsedSizePersent int64
	MountPoint        string
}

var (
	BlacklistedDisks = make(map[string]struct{})
	DiskMaxSize      = make(map[string]int)
	ProcessingDisks  = make(map[string]struct{}) // processingDisks keeps track of disks currently being processed to avoid duplicate operations
)
