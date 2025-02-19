package pvproc

import (
	"context"
	"fmt"
	"log"
	"scully/internal/k8s"
	"scully/internal/metrics"
	"scully/internal/utils"
	"scully/internal/vars"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// processDisk starts a goroutine to monitor a disk's usage after a delay, increasing its size if necessary
func ProcessDisk(clientset *kubernetes.Clientset, metric vars.DiskMetrics, checkDelay time.Duration, threshold, increasePercent int) {

	go func() {
		defer delete(vars.ProcessingDisks, metric.PVName) // Ensure removal from processing map after completion

		pvcName, pvcNamespace, err := k8s.GetPVCByPVName(clientset, metric.PVName)
		if err != nil {
			log.Printf("error when getting PVC name from PV: %s, %v", metric.PVName, err)
		}
		log.Printf("Disk is above %d%% usage, scheduling check. PV name: %s, PVC name: %s, Namespace: %s",
			threshold, metric.PVName, pvcName, pvcNamespace)

		time.Sleep(checkDelay) // Wait checkDelay seconds before re-checking the disk
		updatedMetrics, err := metrics.CollectMetricsFromProc()
		if err != nil {
			log.Printf("Error collecting disk metrics: %v", err)
			return
		}
		// Check if the disk still exceeds threshold% usage after checkDelay seconds
		for _, updatedMetric := range updatedMetrics {
			if updatedMetric.PVName == metric.PVName && updatedMetric.PVUsedSizePersent > int64(threshold) {
				log.Printf("Increasing disk size for %d%%. PV name: %s, PVC name: %s, Namespace: %s",
					increasePercent, metric.PVName, pvcName, pvcNamespace)
				err := increasePVCSize(clientset, metric.PVName, pvcName, pvcNamespace, increasePercent)
				if err != nil {
					log.Printf("Error occurred while increasing PVC size: %v", err)
				}
				// Wait 50 seconds before checking again
				time.Sleep(50 * time.Second)
				return
			}
		}
		log.Printf("Disk usage reduced below %d%%, no action taken. PV name: %s, PVC name: %s, Namespace: %s",
			threshold, metric.PVName, pvcName, pvcNamespace)
	}()
}

func increasePVCSize(clientset *kubernetes.Clientset, pvName, pvcName, pvcNamespace string, increasePercentage int) error {
	pvMaxSizeGb := getPVMaxSize(pvcNamespace)
	pvMaxSizeBytes := int64(pvMaxSizeGb) * 1024 * 1024 * 1024

	currentPvSize, err := k8s.GetPVSizeFromPV(clientset, pvName)
	if err != nil {
		return fmt.Errorf("error when getting the PV size: %s, %v", pvName, err)
	}

	newPVCSize := currentPvSize.DeepCopy()
	newPVCSize.Set(int64(float64(currentPvSize.Value()) * (1 + float64(increasePercentage)/100)))

	if int64(newPVCSize.Value()) > pvMaxSizeBytes {
		log.Printf("New size %d bytes exceeds the max allowed size %d bytes. Skipping expansion. PV name: %s, PVC name: %s, Namespace: %s",
			newPVCSize.Value(), pvMaxSizeBytes, pvName, pvcName, pvcNamespace)
		return nil
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(pvcNamespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error when getting PVC %s: %v", pvcName, err)
	}

	pvcCopy := pvc.DeepCopy()
	pvcCopy.Spec.Resources.Requests[v1.ResourceStorage] = newPVCSize

	_, err = clientset.CoreV1().PersistentVolumeClaims(pvcNamespace).Update(context.TODO(), pvcCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error when updating PVC size %s: %v", pvcName, err)
	}

	log.Printf("Disk size has been increased by %d%%. New size: %d bytes, PV name: %s, PVC name: %s, Namespace: %s\n",
		increasePercentage, newPVCSize.Value(), pvName, pvcName, pvcNamespace)
	return nil

}

func getPVMaxSize(pvcNamespace string) int {
	maxSizeGB, exists := vars.DiskMaxSize[pvcNamespace]
	if !exists {
		defaultMaxSize := utils.GetDiskMaxSizeGb()
		log.Printf("Namespace %s not found in DiskMaxSizeMap. Using default value: %d GB", pvcNamespace, defaultMaxSize)
		return defaultMaxSize
	}

	return maxSizeGB
}
