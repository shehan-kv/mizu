package v1

import (
	"encoding/json"
	"mizu/internal/db/store"
	"mizu/internal/dto/file"
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
	mux.Handle("GET /{channelId}", mwChain.Handle(fileHndl.GetByChannel))
	mux.Handle("GET /project/{projectId}", mwChain.Handle(fileHndl.GetByProject))

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

func (fileHndl *FileHandler) GetByChannel(w http.ResponseWriter, r *http.Request) {

	channelId := r.PathValue("channelId")
	parsedChId, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil || parsedChId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keyword := r.URL.Query().Get("q")
	strPage := r.URL.Query().Get("page")
	strLimit := r.URL.Query().Get("limit")

	var page int64
	var limit int64

	if strPage == "" {
		page = 1
	} else {
		parsedPage, err := strconv.ParseInt(strPage, 10, 64)
		if err != nil || parsedPage <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page = parsedPage
	}

	if strLimit == "" {
		limit = 15
	} else {
		parsedLimit, err := strconv.ParseInt(strLimit, 10, 64)
		if err != nil || parsedLimit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	result, err := fileHndl.fileSrv.GetByChannelId(r.Context(), parsedChId, &file.FileSearch{
		Keyword: keyword,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (fileHndl *FileHandler) GetByProject(w http.ResponseWriter, r *http.Request) {

	prjId := r.PathValue("projectId")
	parsedPrjId, err := strconv.ParseInt(prjId, 10, 64)
	if err != nil || parsedPrjId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keyword := r.URL.Query().Get("q")
	strPage := r.URL.Query().Get("page")
	strLimit := r.URL.Query().Get("limit")

	var page int64
	var limit int64

	if strPage == "" {
		page = 1
	} else {
		parsedPage, err := strconv.ParseInt(strPage, 10, 64)
		if err != nil || parsedPage <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		page = parsedPage
	}

	if strLimit == "" {
		limit = 15
	} else {
		parsedLimit, err := strconv.ParseInt(strLimit, 10, 64)
		if err != nil || parsedLimit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	result, err := fileHndl.fileSrv.GetByProjectId(r.Context(), parsedPrjId, &file.FileSearch{
		Keyword: keyword,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
