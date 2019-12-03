// +build !windows,!darwin

package pinger

import (
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
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|syscall.O_DIRECT, 0755)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	file.Write(data)

	return time.Since(start).Microseconds(), nil
}
