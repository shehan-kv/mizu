package v1

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/session"
	"mizu/internal/sse"
	"net/http"
)

// Handles SSE related requests
//
// Uses an SseSender to to perform
// SSE operations
type SseHandler struct {
	sseSndr *sse.SseSender
}

// Creates a new instance of SseHandler
//
// Parameters:
//   - sseSndr: a pointer to a sse.SseSender
//
// Returns:
//   - a pointer to a new SseHandler
func NewSseHandler(sseSndr *sse.SseSender) *SseHandler {
	return &SseHandler{
		sseSndr: sseSndr,
	}
}

// Creates a ServeMux for the SSE routes and middleware.
// Defines the routes and handler function for each route.
// Registers middleware for the routes.
//
// Parameters:
//   - lg: an implementation of logger.Logger
//   - seSt: an implementation of session.SessionStore
//   - usrSt: an implementation of store.UserStore
//
// Returns:
//   - a *http.ServeMux
func (sseHndl *SseHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("GET /", mwChain.Handle(sseHndl.Events))

	return mux
}

// Handles SSE client connection and streams events
//
// Method: GET
//
// Possible Response Codes:
//   - 500 InternalServerError - Server error
func (sseHndl *SseHandler) Events(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := middleware.GetUserFromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch := sseHndl.sseSndr.AddClient(r.RemoteAddr, user.Id)

	disconnected := r.Context().Done()

	for {
		select {
		case <-disconnected:
			sseHndl.sseSndr.RemoveClient(r.RemoteAddr, user.Id)
			return

		case msg := <-ch:
			w.Write(msg)
			flusher.Flush()
		}

	}
}
