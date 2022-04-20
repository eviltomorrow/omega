package lock

type FileLock interface {
	Path() string
	Release() error
}
