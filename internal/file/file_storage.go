package file

type FileStorage interface {
	Store() error

	Close() error
}
