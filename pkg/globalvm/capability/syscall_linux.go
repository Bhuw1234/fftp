//go:build linux

package capability

import (
	"syscall"
)

// syscallStatfs is the statfs structure for this platform.
type syscallStatfs = syscall.Statfs_t

// statfs calls the system statfs.
func statfs(path string, stat *syscallStatfs) error {
	return syscall.Statfs(path, stat)
}
