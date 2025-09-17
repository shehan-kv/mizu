package file

import (
	"mizu/internal/logger"
	"os"
	"strings"
)

func GetFileStorage(lg logger.Logger) FileStorage {

	fileStorage := strings.TrimSpace(strings.ToLower(os.Getenv("FILE_STORAGE")))

	switch fileStorage {
	case "disk":
		lg.Info("Disk file storage selected")
		return NewDiskFileStorage(lg)

	default:
		lg.Info("Storage not specified")
		lg.Info("Using disk file storage with default settings")
		return NewDiskFileStorage(lg)
	}
}
