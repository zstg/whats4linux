//go:build windows

package lockfile

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "syscall"
    "unsafe"
)

type Lock struct {
    file *os.File
    path string
}

var (
    kernel32         = syscall.NewLazyDLL("kernel32.dll")
    procLockFileEx   = kernel32.NewProc("LockFileEx")
    procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
    LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
    LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
)

func Acquire(lockPath string) (*Lock, error) {
    absPath, err := filepath.Abs(lockPath)
    if err != nil {
        return nil, fmt.Errorf("[ERROR] resolve lock path: %w", err)
    }

    f, err := os.OpenFile(absPath, os.O_CREATE|os.O_RDWR, 0644)
    if err != nil {
        return nil, fmt.Errorf("[ERROR] open lock file: %w", err)
    }

    // Lock the file using Windows LockFileEx
    var overlapped syscall.Overlapped
    flags := LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY
    
    r1, _, err := procLockFileEx.Call(
        uintptr(f.Fd()),
        uintptr(flags),
        0,
        1,
        0,
        uintptr(unsafe.Pointer(&overlapped)),
    )
    
    if r1 == 0 {
        f.Close()
        return nil, fmt.Errorf("[ERROR] (%s) another instance of the app is already running or lock failed: %w", absPath, err)
    }

    _ = f.Truncate(0)
    _, _ = fmt.Fprintf(f, "%d\n", os.Getpid())

    return &Lock{
        file: f,
        path: absPath,
    }, nil
}

func (l *Lock) Release() error {
    if l == nil || l.file == nil {
        return nil
    }

    // Unlock the file
    var overlapped syscall.Overlapped
    procUnlockFileEx.Call(
        uintptr(l.file.Fd()),
        0,
        1,
        0,
        uintptr(unsafe.Pointer(&overlapped)),
    )

    err := l.file.Close()
    _ = os.Remove(l.path)
    return err
}

func EnsureSingleInstance(lockPath string) *Lock {
    lock, err := Acquire(lockPath)
    if err != nil {
        log.Fatal(err)
    }
    return lock
}