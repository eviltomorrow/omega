//go:build darwin
// +build darwin

package lock

import (
	"os"
	"strconv"
	"syscall"
)

type unixFileLock struct {
	p string
	f *os.File
}

func (fl *unixFileLock) Release() error {
	if err := setFileLock(fl.f, false, false); err != nil {
		return err
	}
	return fl.f.Close()
}

func (fl *unixFileLock) Path() string {
	return fl.p
}

func newFileLock(path string, readOnly bool) (fl FileLock, err error) {
	var flag int
	if readOnly {
		flag = os.O_RDONLY
	} else {
		flag = os.O_RDWR
	}
	f, err := os.OpenFile(path, flag, 0)
	if os.IsNotExist(err) {
		f, err = os.OpenFile(path, flag|os.O_CREATE, 0644)
	}
	if err != nil {
		return
	}
	err = setFileLock(f, readOnly, true)
	if err != nil {
		f.Close()
		return
	}
	if !readOnly {
		if _, err = f.Write([]byte(strconv.Itoa(os.Getpid()))); err != nil {
			return
		}
		if err = f.Sync(); err != nil {
			return
		}
	}
	fl = &unixFileLock{p: path, f: f}
	return
}

func setFileLock(f *os.File, readOnly, lock bool) error {
	how := syscall.LOCK_UN
	if lock {
		if readOnly {
			how = syscall.LOCK_SH
		} else {
			how = syscall.LOCK_EX
		}
	}
	return syscall.Flock(int(f.Fd()), how|syscall.LOCK_NB)
}
