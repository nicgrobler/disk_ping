// +build windows

package pinger

import (
	"os"
	"syscall"
	"time"
)

const (
	FILE_FLAG_NO_BUFFERING  = 0x20000000
	FILE_FLAG_WRITE_THROUGH = 0x80000000
)

/*

original source from Go https://golang.org/src/syscall/syscall_windows.go

if len(path) == 0 {
		return InvalidHandle, ERROR_FILE_NOT_FOUND
	}
	pathp, err := UTF16PtrFromString(path)
	if err != nil {
		return InvalidHandle, err
	}
	var access uint32
	switch mode & (O_RDONLY | O_WRONLY | O_RDWR) {
	case O_RDONLY:
		access = GENERIC_READ
	case O_WRONLY:
		access = GENERIC_WRITE
	case O_RDWR:
		access = GENERIC_READ | GENERIC_WRITE
	}
	if mode&O_CREAT != 0 {
		access |= GENERIC_WRITE
	}
	if mode&O_APPEND != 0 {
		access &^= GENERIC_WRITE
		access |= FILE_APPEND_DATA
	}
	sharemode := uint32(FILE_SHARE_READ | FILE_SHARE_WRITE)
	var sa *SecurityAttributes
	if mode&O_CLOEXEC == 0 {
		sa = makeInheritSa()
	}
	var createmode uint32
	switch {
	case mode&(O_CREAT|O_EXCL) == (O_CREAT | O_EXCL):
		createmode = CREATE_NEW
	case mode&(O_CREAT|O_TRUNC) == (O_CREAT | O_TRUNC):
		createmode = CREATE_ALWAYS
	case mode&O_CREAT == O_CREAT:
		createmode = OPEN_ALWAYS
	case mode&O_TRUNC == O_TRUNC:
		createmode = TRUNCATE_EXISTING
	default:
		createmode = OPEN_EXISTING
	}
	h, e := CreateFile(pathp, access, sharemode, sa, createmode, FILE_ATTRIBUTE_NORMAL, 0)
	return h, e
*/

// UnbufferedWriteTime writes data directly to the underlying storage, bypassing caching and
// giving a "time taken" in microseconds
func UnbufferedWriteTime(path string, data []byte) (int64, error) {
	// here we use syscall to enable us to pass certain flags to the OS to disable caching
	// this means that we are measuring "true" disk latency, and not, the latency of writing to fs cache
	start := time.Now()

	if len(path) == 0 {
		return InvalidHandle, ERROR_FILE_NOT_FOUND
	}
	pathp, err := UTF16PtrFromString(path)
	if err != nil {
		return InvalidHandle, err
	}

	var access uint32
	access = GENERIC_READ | GENERIC_WRITE
	access |= GENERIC_WRITE

	sharemode := uint32(syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE)
	var sa *syscall.SecurityAttributes
	var createmode uint32
	createmode = CREATE_ALWAYS

	h, e := syscall.CreateFile(&pathp[0], access, sharemode, sa, createmode, syscall.FILE_ATTRIBUTE_NORMAL|FILE_FLAG_NO_BUFFERING|FILE_FLAG_WRITE_THROUGH, 0)
	if e != nil {
		return nil, &os.PathError{"create", path, e}
	}
	file := os.NewFile(uintptr(h), path)
	defer file.Close()

	file.Write(data)

	return time.Since(start).Microseconds(), nil
}
