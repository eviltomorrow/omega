package lock

import "os"

func CreateFileLock(path string) (FileLock, error) {
	lock, err := newFileLock(path, false)
	if err != nil {
		return nil, err
	}
	return lock, nil
}

func DestroyFileLock(lock FileLock) error {
	if err := lock.Release(); err != nil {
		return err
	}
	return os.Remove(lock.Path())
}
