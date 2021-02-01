package infrastructure

import (
	"os"
	"path/filepath"
)

//GetVolumeSize returns size of directory in MB
func GetDirectorySize(path string) int64 {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	if err != nil {
		return -1
	}

	return size / 1024 / 1024
}
