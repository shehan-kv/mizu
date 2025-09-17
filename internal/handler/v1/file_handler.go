package v1

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileHandler provides HTTP handlers
// for file related endpoints.
type FileHandler struct {
	fileSrv   *service.FileService
	maxMemory int64
}

// NewFileHandler constructs a new FileHandler.
func NewFileHandler(fileSrv *service.FileService) *FileHandler {
	var maxMemory int64
	strMaxMemory := strings.TrimSpace(os.Getenv("FILE_MAX_MEMORY"))
	parsedMemory, err := strconv.Atoi(strMaxMemory)

	if err != nil {
		maxMemory = 50 << 20
	} else {
		maxMemory = int64(parsedMemory) << 20
	}

	return &FileHandler{
		fileSrv:   fileSrv,
		maxMemory: maxMemory,
	}
}

// GetMux returns an http.ServeMux for the file routes and middleware.
// It defines the routes and handler function for each route.
// Registers middleware for the routes.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//   - seSt: an implementation of session.SessionStore
//   - usrSt: an implementation of store.UserStore
//
// Returns:
//   - a *http.ServeMux
func (fileHndl *FileHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /", mwChain.Handle(fileHndl.StoreFile))

	return mux
}

func (fileHndl *FileHandler) StoreFile(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(fileHndl.maxMemory); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if filepath.Base(header.Filename) == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer file.Close()

	channelId := r.FormValue("channelId")
	parsedChId, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil || parsedChId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = fileHndl.fileSrv.StoreFile(r.Context(), parsedChId, file, header)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
