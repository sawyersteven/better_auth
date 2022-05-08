package files

import (
	"io/fs"
	"os"
	"path/filepath"
)

/// Opens a file, attempting to create the neccessary dir tree if it does not exist
func MkDirsAndOpen(filePath string, flag int, perm fs.FileMode) (*os.File, error) {
	flag |= os.O_CREATE

	// In case dir tree doesn't exist, try to make it
	dir, _ := filepath.Split(filePath)
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filePath, flag, perm)
}

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
