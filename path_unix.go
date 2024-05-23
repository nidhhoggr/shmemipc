//go:build !windows

package shmemipc

import "path/filepath"

func getFilename(name string) string {
	return filepath.Join("/tmp", name)
}
