package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	defaultDiskMaxSizeGb   = 100
	defaultDelaySec        = 60
	defaultThreshold       = 85
	defaultIncreasePercent = 10
	defaultBlackListConfig = "pv-blacklist"
	defaultLimitsConfig    = "pv-max-size"
)

func GetDiskMaxSizeGb() int {
	envValue := os.Getenv("DISK_MAX_SIZE_GB")
	if envValue == "" {
		log.Printf("Environment variable DISK_MAX_SIZE_GB is not set. Using default value: %d GB", defaultDiskMaxSizeGb)
		return defaultDiskMaxSizeGb
	}

	maxSizeGB, err := strconv.Atoi(envValue)
	if err != nil {
		log.Printf("Invalid value '%s' for DISK_MAX_SIZE_GB. Using default value: %d GB", envValue, defaultDiskMaxSizeGb)
		return defaultDiskMaxSizeGb
	}

	return maxSizeGB
}

func GetDiskCheckDelay() time.Duration {
	delayStr := os.Getenv("DISK_CHECK_DELAY")
	if delayStr == "" {
		return defaultDelaySec * time.Second
	}
	delay, err := strconv.Atoi(delayStr)
	if err != nil || delay < 30 {
		log.Printf("Invalid DISK_CHECK_DELAY value: %s, using default 60 seconds", delayStr)
		return defaultDelaySec * time.Second
	}
	return time.Duration(delay) * time.Second
}

func GetDiskUsageThreshold() int {
	thresholdStr, exists := os.LookupEnv("DISK_USAGE_THRESHOLD")
	if !exists {
		return defaultThreshold
	}

	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil || threshold < 30 || threshold > 90 {
		log.Printf("Invalid DISK_USAGE_THRESHOLD value: %s, using default %d%%", thresholdStr, defaultThreshold)
		return defaultThreshold
	}

	return threshold
}

func GetDiskIncreasePercentage() int {
	increasePercentStr, exists := os.LookupEnv("DISK_INCREASE_PERCENTAGE")
	if !exists {
		return defaultIncreasePercent
	}

	increasePercent, err := strconv.Atoi(increasePercentStr)
	if err != nil || increasePercent <= 0 || increasePercent > 50 {
		log.Printf("Invalid DISK_INCREASE_PERCENTAGE value: %s, using default %d%%", increasePercentStr, defaultIncreasePercent)
		return defaultIncreasePercent
	}

	return increasePercent
}

func GetBlackListedDiskConfigMapName() string {
	configMapName := os.Getenv("DISK_BLACKLIST_CONFIGMAP")
	if configMapName == "" {
		configMapName = defaultBlackListConfig
	}

	return configMapName
}

func GetBlackListedDiskConfigMapNamespace() string {
	namespace := os.Getenv("DISK_BLACKLIST_NAMESPACE")
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			log.Fatalf("Error: Neither DISK_BLACKLIST_NAMESPACE nor POD_NAMESPACE environment variable is set")
		}
	}

	return namespace
}

func GetDiskMaxSizeConfigMapName() string {
	configMapName := os.Getenv("DISK_MAX_SIZE_CONFIGMAP")
	if configMapName == "" {
		configMapName = defaultLimitsConfig
	}

	return configMapName
}

func GetDiskMaxSizeConfigMapNamespace() string {
	namespace := os.Getenv("DISK_MAX_SIZE_NAMESPACE")
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			log.Fatalf("Error: Neither DISK_MAX_SIZE_NAMESPACE nor POD_NAMESPACE environment variable is set")
		}
	}

	return namespace
}
