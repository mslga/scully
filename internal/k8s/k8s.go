package k8s

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"scully/internal/vars"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Create a client to connect to Kubernetes API
func CreateKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error when connecting to Kubernetes: %v", err)
	}

	// Create a client to interact with Kubernetes API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error when creating a Kubernetes client: %v", err)
	}

	return clientset, nil
}

func GetPVCByPVName(clientset *kubernetes.Clientset, pvName string) (string, string, error) {
	// Get PVC name related to PV
	pvcName := ""
	pvcNamespace := ""

	// Find PVC according to PV
	pvcList, err := clientset.CoreV1().PersistentVolumeClaims("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", "", fmt.Errorf("error when obtaining the PVC list: %v", err)
	}

	for _, pvc := range pvcList.Items {
		if pvc.Spec.VolumeName == pvName {
			pvcName = pvc.Name
			pvcNamespace = pvc.Namespace
			break
		}
	}

	if pvcName == "" {
		return "", "", fmt.Errorf("PVC associated with PV %s not found", pvName)
	} else {
		return pvcName, pvcNamespace, nil
	}
}

func GetPVSizeFromPV(clientset *kubernetes.Clientset, pvName string) (resource.Quantity, error) {
	var nilRQ resource.Quantity
	pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pvName, metav1.GetOptions{})
	if err != nil {
		return nilRQ, fmt.Errorf("error when getting a PV with the name %s: %v", pvName, err)
	}

	// Calculating the size of PV
	currentSizePV := pv.Spec.Capacity["storage"]

	return currentSizePV, nil
}

// LoadBlacklistedDisks reads the ConfigMap and populates the blacklistedDisks map with PVC names
// from all keys in the ConfigMap, checking that the values start with 'pvc-'.
func LoadBlacklistedDisks(clientset *kubernetes.Clientset, namespace, configMapName string) {
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error fetching ConfigMap %s in namespace %s: %v", configMapName, namespace, err)
	}

	// Iterate over all keys in the ConfigMap
	for _, blacklistData := range configMap.Data {
		// Split the list of PVCs into separate entries
		pvList := strings.Split(blacklistData, "\n")
		for _, pv := range pvList {
			// Check if the PVC name starts with 'pvc-'
			if strings.HasPrefix(pv, "pvc-") && pv != "" {
				vars.BlacklistedDisks[pv] = struct{}{}
			}
		}
	}

	var logEntries []string
	for pv := range vars.BlacklistedDisks {
		pvcName, pvcNamespace, err := GetPVCByPVName(clientset, pv)
		if err != nil {
			log.Printf("Warning: Could not find PVC for PV %s: %v", pv, err)
			logEntries = append(logEntries, fmt.Sprintf("PV Name: %s (PVC not found)", pv))
		} else {
			logEntries = append(logEntries, fmt.Sprintf("PV Name: %s, PVC Name: %s, Namespace: %s", pv, pvcName, pvcNamespace))
		}
	}

	log.Printf("Total %d blacklisted disks loaded from ConfigMap %s", len(vars.BlacklistedDisks), configMapName)

	if len(logEntries) > 0 {
		log.Printf("Blacklisted PV mappings: %s", strings.Join(logEntries, "; "))
	}

}

func LoadDiskMaxSize(clientset *kubernetes.Clientset, namespace, configMapName string) {
	configMap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Error fetching ConfigMap %s in namespace %s: %v", configMapName, namespace, err)
	}

	// Clean old map entries
	vars.DiskMaxSize = make(map[string]int)

	for ns, sizeStr := range configMap.Data {
		sizeGB, err := strconv.Atoi(sizeStr)
		if err != nil {
			log.Printf("Invalid size '%s' for namespace %s in ConfigMap %s. Skipping.", sizeStr, ns, configMapName)
			continue
		}
		vars.DiskMaxSize[ns] = sizeGB
	}

	log.Printf("Loaded %d entries for disk max size limits from ConfigMap %s", len(vars.DiskMaxSize), configMapName)

	for ns, size := range vars.DiskMaxSize {
		log.Printf("Namespace: %s, Max Size Limit: %dGB", ns, size)
	}
}
