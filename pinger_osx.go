// +build darwin

package pinger

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

// UnbufferedWriteTime writes data directly to the underlying storage, bypassing caching and
// giving a "time taken" in microseconds
func UnbufferedWriteTime(filename string, data []byte) (int64, error) {
	// here we use syscall to enable us to pass certain flags to the OS to disable caching
	// this means that we are measuring "true" disk latency, and not, the latency of writing to fs cache
	start := time.Now()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Set F_NOCACHE to avoid caching
	// F_NOCACHE    Turns data caching off/on. A non-zero value in arg turns data caching off.  A value
	//              of zero in arg turns data caching on.
	_, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, uintptr(file.Fd()), syscall.F_NOCACHE, 1)
	if e1 != 0 {
		err = fmt.Errorf("Failed to set F_NOCACHE: %s", e1)
		file.Close()
		file = nil
		return 0, err
	}

	file.Write(data)

	return time.Since(start).Microseconds(), nil
}
