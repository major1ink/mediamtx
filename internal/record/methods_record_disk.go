package record

import (
	"syscall"
)

type DiskStatus struct {
	// всего места на диске
	All uint64
	// используемое пространство
	used uint64
	// свободное пространство
	free uint64
	// доступное пространство
	Avail uint64
}

func diskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}

	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Avail = fs.Bavail * uint64(fs.Bsize)
	disk.used = disk.All - disk.free
	return
}

func getPaths(drives []interface{}) []string {
	paths := []string{}
	for _, p := range drives {
		paths = append(paths, p.(string))
	}

	return paths
}

func allDiskUsages(drives []interface{}) map[string]float64 {

	infoDisks := make(map[string]float64)
	paths := getPaths(drives)

	for _, path := range paths {
		disk := diskUsage(path)
		infoDisks[path] = (float64(disk.All) - float64(disk.Avail)) / float64(disk.All)
	}

	return infoDisks
}

func getMostFreeDisk(drives []interface{}) (mostFree string) {

	var min float64
	infoDisks := allDiskUsages(drives)
	
	for _, min = range infoDisks {
		break
	}

	for disk, free := range infoDisks {
		if free > min {
			continue
		}

		mostFree, min = disk, free
	}

	return mostFree
}

func getMostFreeDiskGroup(drives map[string]string) (mostFree string) {

	var min float64
	infoDisks := allDiskUsagesGroup(drives)
	for _, min = range infoDisks {
		break
	}

	for disk, free := range infoDisks {
		if free > min {
			continue
		}

		mostFree, min = disk, free
	}

	return mostFree
}

func allDiskUsagesGroup(drives map[string]string) map[string]float64 {

	infoDisks := make(map[string]float64)

	for path := range drives {
		disk := diskUsage(path)
		infoDisks[path] = (float64(disk.All) - float64(disk.Avail)) / float64(disk.All)
	}
	return infoDisks
}
