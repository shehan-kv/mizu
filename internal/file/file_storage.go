package file

type FileStorage interface {
	UploadFile() error

	Close() error
}
