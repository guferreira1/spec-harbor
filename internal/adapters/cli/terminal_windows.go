//go:build windows

package cli

import (
	"os"
	"syscall"
)

func isInputTerminalFile(file *os.File) bool {
	if file == nil {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	if stat.Mode()&os.ModeCharDevice == 0 {
		return false
	}

	var mode uint32
	err = syscall.GetConsoleMode(syscall.Handle(file.Fd()), &mode)
	return err == nil
}
