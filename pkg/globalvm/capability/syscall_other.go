//go:build !linux

package capability

// syscallStatfs is a dummy for non-linux platforms.
type syscallStatfs struct {
	Blocks  uint64
	Bsize   int64
	Bavail  uint64
}

// statfs is a stub for non-linux platforms.
func statfs(path string, stat *syscallStatfs) error {
	// Return a reasonable default for non-linux platforms
	stat.Blocks = 1000000
	stat.Bsize = 4096
	stat.Bavail = 800000
	return nil
}
