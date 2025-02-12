# Scully

Utility to dynamically increase pv size depending on disk utilization

## Description

Metrics are taken from procfs file system and statfs system call every 10 seconds for disks with **pvc-** prefix.
If the disk size exceeds the specified treshold (DISK_USAGE_THRESHOLD), the disk is additionally checked for some more time (DISK_CHECK_DELAY) and if its size does not decrease, the disk is increased by a certain size (DISK_INCREASE_PERCENTAGE).

## Requirements

* Kubernetes version >= 1.29.0
* CSI driver should support dynamic pv increase and automatically increase FS size in pod, because the utility does not restart pods after dynamic resizing
* In storageclass, *allowVolumeExpansion* flag should be `true`

```yaml
allowVolumeExpansion: true
```

## Helm chart

Install scully with helm chart

```bash
helm upgrade --install scully helm/ -f helm/values.yaml --namespace scully --create-namespace
```
