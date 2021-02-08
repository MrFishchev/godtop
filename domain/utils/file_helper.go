package utils

import (
	"os"
	"path/filepath"
)

//GetVolumeSize returns size of directory in MB
func GetDirectorySize(path string) uint64 {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return err
	})

	if err != nil {
		return 0
	}

	return size
}
