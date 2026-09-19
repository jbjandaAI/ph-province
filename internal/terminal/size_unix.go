//go:build linux || darwin

package terminal

import (
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

type windowSize struct {
	rows    uint16
	columns uint16
	xpixel  uint16
	ypixel  uint16
}

func nativeSize(file *os.File) (int, int, bool) {
	request := uintptr(0x5413) // Linux TIOCGWINSZ.
	if runtime.GOOS == "darwin" {
		request = 0x40087468 // Darwin TIOCGWINSZ.
	}
	var size windowSize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), request, uintptr(unsafe.Pointer(&size)))
	if errno != 0 {
		return 0, 0, false
	}
	return int(size.columns), int(size.rows), true
}
