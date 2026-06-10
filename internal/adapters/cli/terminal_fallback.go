//go:build !linux && !darwin && !freebsd && !netbsd && !openbsd && !dragonfly && !windows

package cli

import "os"

func isInputTerminalFile(file *os.File) bool {
	if file == nil {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}
