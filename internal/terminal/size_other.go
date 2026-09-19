//go:build !linux && !darwin

package terminal

import "os"

func nativeSize(_ *os.File) (int, int, bool) {
	return 0, 0, false
}
