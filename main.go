package main

import (
	"log"
	"time"

	"scully/internal/k8s"
	"scully/internal/metrics"
	"scully/internal/pvproc"
	"scully/internal/utils"
	"scully/internal/vars"
)

func main() {
	// Create a Kubernetes client
	clientset, err := k8s.CreateKubernetesClient()
	if err != nil {
		log.Fatalf("Error while creating Kubernetes client: %v", err)
	}

	diskCheckDelay := utils.GetDiskCheckDelay()
	diskMaxSize := utils.GetDiskMaxSizeGb()
	threshold := utils.GetDiskUsageThreshold()
	increasePercent := utils.GetDiskIncreasePercentage()
	blackListedDiskConfigMapName := utils.GetBlackListedDiskConfigMapName()
	blackListedDiskConfigMapNamespace := utils.GetBlackListedDiskConfigMapNamespace()
	diskMaxSizeConfigMapName := utils.GetDiskMaxSizeConfigMapName()
	diskMaxSizeConfigMapNamespace := utils.GetDiskMaxSizeConfigMapNamespace()
	log.Printf("Scully settings: DISK_CHECK_DELAY=%v DISK_MAX_SIZE_GB=%v DISK_USAGE_THRESHOLD=%v%% DISK_INCREASE_PERCENTAGE=%v%%",
		diskCheckDelay, diskMaxSize, threshold, increasePercent)

	k8s.LoadBlacklistedDisks(clientset, blackListedDiskConfigMapNamespace, blackListedDiskConfigMapName) // Load the blacklisted disks from the ConfigMap
	k8s.LoadDiskMaxSize(clientset, diskMaxSizeConfigMapNamespace, diskMaxSizeConfigMapName)              // Load disk max size limits from the ConfigMap

	for {
		// Collect disk usage metrics from /proc/mounts
		m, err := metrics.CollectMetricsFromProc()
		if err != nil {
			log.Printf("Error collecting disk metrics: %v", err)
			continue
		}

		for _, metric := range m {
			if _, blacklisted := vars.BlacklistedDisks[metric.PVName]; blacklisted {
				continue
			}

			// Check if the disk is already being processed
			if _, exists := vars.ProcessingDisks[metric.PVName]; exists {
				continue
			}

			// Process disks that exceed threshold usage
			if metric.PVUsedSizePersent > int64(threshold) {
				// Mark the disk as being processed
				vars.ProcessingDisks[metric.PVName] = struct{}{}
				pvproc.ProcessDisk(clientset, metric, diskCheckDelay, threshold, increasePercent)
			}
		}
		// Wait 10 seconds before checking again
		time.Sleep(10 * time.Second)
	}
}
