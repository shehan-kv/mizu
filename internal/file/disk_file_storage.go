package file

import (
	"errors"
	"io"
	"io/fs"
	"mime/multipart"
	"mizu/internal/logger"
	"os"
	"path"
	"path/filepath"
)

type DiskFileStorage struct {
	baseDirectory string
	fileDirectory string
	userDirectory string
}

func NewDiskFileStorage(lg logger.Logger) *DiskFileStorage {

	storageDirectory := os.Getenv("STORAGE_DIRECTORY")

	if len(storageDirectory) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			lg.Fatal("Could not determine current directory")
		}

		storageDirectory = path.Join(wd, "storage")
	}

	if _, err := os.Stat(storageDirectory); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			lg.Warn("Storage directory does not exist")

			if err := os.Mkdir(storageDirectory, 0700); err != nil {
				lg.Fatal("Could not create storage directory")
			}

			lg.Info("Storage directory created")

		} else {
			lg.Fatal("Unexpected error occured with storage")
		}
	}

	if !isWritable(storageDirectory) {
		lg.Fatal("Cannot write files to storage directory")
	}

	fileDirectory := path.Join(storageDirectory, "file")
	if _, err := os.Stat(fileDirectory); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			lg.Warn("File storage directory does not exist")

			if err := os.Mkdir(fileDirectory, 0700); err != nil {
				lg.Fatal("Could not create file storage directory")
			}

			lg.Info("File storage directory created")

		} else {
			lg.Fatal("Unexpected error occured with file storage")
		}
	}

	userDirectory := path.Join(storageDirectory, "user")
	if _, err := os.Stat(userDirectory); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			lg.Warn("User storage directory does not exist")

			if err := os.Mkdir(userDirectory, 0700); err != nil {
				lg.Fatal("Could not create user storage directory")
			}

			lg.Info("User storage directory created")

		} else {
			lg.Fatal("Unexpected error occured with user storage")
		}

	}

	return &DiskFileStorage{
		baseDirectory: storageDirectory,
		fileDirectory: fileDirectory,
		userDirectory: userDirectory,
	}
}

func isWritable(path string) bool {
	tmpFile := "tmpfile"
	file, err := os.CreateTemp(path, tmpFile)
	if err != nil {
		return false
	}
	defer os.Remove(file.Name())
	defer file.Close()

	return true
}

func (u *DiskFileStorage) Store(fileType FileType, file multipart.File, fileName string) (string, error) {

	var fPath string

	switch fileType {
	case TypeFile:
		fPath = u.fileDirectory

	case TypeUser:
		fPath = u.userDirectory
	}

	tmp, err := os.CreateTemp(fPath, "upload-*.tmp")
	if err != nil {
		return "", ErrFileCreateFailed
	}

	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", ErrCopyFailed
	}

	_ = tmp.Sync()

	if err := tmp.Close(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", ErrFileCloseFailed
	}

	// Owner read/write only permissions
	_ = os.Chmod(tmpPath, 0o600)

	finalPath := filepath.Join(fPath, fileName)

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", ErrRenameFailed
	}

	// The URL of the file.
	// Server should respond to this path
	fileUrl := "/api/v1/storage/files/" + fileName

	return fileUrl, nil
}

func (u *DiskFileStorage) Remove(url string) error {

	if err := os.Remove(url); err != nil {
		return ErrRemoveFailed
	}

	return nil
}

func (u *DiskFileStorage) Close() error {
	return nil
}
