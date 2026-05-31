package sse

import (
	"mizu/internal/presentation/http/rest/middleware"
	"net/http"
)

// Handler handles SSE related requests
type Handler struct {
	sender *Sender
}

func NewHandler(sender *Sender) *Handler {
	return &Handler{
		sender: sender,
	}
}

func (h *Handler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {

	mux := http.NewServeMux()

	mux.Handle("GET /", authMiddleware(http.HandlerFunc(h.Events)))

	return mux
}

// Events handles SSE client connection and streams events
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	actorID := middleware.ActorIDFromContext(r.Context())
	if actorID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch := h.sender.AddClient(r.RemoteAddr, actorID)

	disconnected := r.Context().Done()

	for {
		select {
		case <-disconnected:
			h.sender.RemoveClient(r.RemoteAddr, actorID)
			return

		case msg := <-ch:
			w.Write(msg)
			flusher.Flush()
		}

	}
}
