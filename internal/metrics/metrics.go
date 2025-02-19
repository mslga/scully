package metrics

import (
	"bufio"
	"fmt"
	"os"
	"scully/internal/vars"
	"strings"
	"syscall"
)

func extractPVName(mountPoint string) string {
	// Extract PVC from the mount path
	// Ex: /var/lib/kubelet/pods/<UUID>/volumes/.../pvc-<UUID>/mount
	parts := strings.Split(mountPoint, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "pvc-") {
			return part
		}
	}

	return ""
}

func CollectMetricsFromProc() ([]vars.DiskMetrics, error) {
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/mounts: %v", err)
	}
	defer file.Close()

	var metrics []vars.DiskMetrics
	uniqueMetrics := make(map[string]struct{}) // Map for tracking unique PVCs

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]

		// Filter only the right devices (/dev/sd[b-z]).
		if strings.HasPrefix(device, "/dev/sd") && !strings.HasPrefix(device, "/dev/sda") {
			// Trying to extract the PV from the mount path
			pvName := extractPVName(mountPoint)
			if pvName == "" {
				continue
			}

			// Form a unique key for verification (e.g. combination of device and PV)
			uniqueKey := fmt.Sprintf("%s_%s", device, pvName)

			// Check if such a PV has already been added
			if _, exists := uniqueMetrics[uniqueKey]; exists {
				continue // If there is already such a PV, we skip it
			}

			// We add to the map that this PV has already been processed
			uniqueMetrics[uniqueKey] = struct{}{}

			// Collecting file system statistics
			var stat syscall.Statfs_t
			if err := syscall.Statfs(mountPoint, &stat); err != nil {
				fmt.Printf("error reading filesystem stats for %s: %v", mountPoint, err)
				continue
			}

			// Calculate total size and used size in KB
			totalSize := stat.Blocks * uint64(stat.Bsize) / 1024 // KB
			usedSize := (stat.Blocks - stat.Bavail) * uint64(stat.Bsize) / 1024

			// Calculate the utilization percentage
			usedPercent := float64(usedSize) / float64(totalSize) * 100

			metrics = append(metrics, vars.DiskMetrics{
				PVName:            pvName,
				PVTotalSize:       int64(totalSize),
				PVUsedSize:        int64(usedSize),
				PVUsedSizePersent: int64(usedPercent),
				MountPoint:        mountPoint,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/mounts: %v", err)
	}

	return metrics, nil
}
