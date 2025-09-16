package v1

import "mizu/internal/service"

// FileHandler provides HTTP handlers
// for file related endpoints.
type FileHandler struct {
	fileSrv *service.FileService
}

// NewFileHandler constructs a new FileHandler.
func NewFileHandler(fileSrv *service.FileService) *FileHandler {
	return &FileHandler{
		fileSrv: fileSrv,
	}
}
