package file

import "mime/multipart"

type FileStorage interface {
	Store(fileType FileType, file multipart.File, fileName string) (string, error)

	Close() error
}
