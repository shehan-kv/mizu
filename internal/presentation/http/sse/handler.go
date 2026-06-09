package sse

import (
	"fmt"
	"mizu/internal/application/logger"
	"mizu/internal/presentation/http/rest/middleware"
	"net/http"
	"time"
)

// Handler handles SSE related requests
type Handler struct {
	sender *Sender
	log    logger.Logger
}

func NewHandler(sender *Sender, log logger.Logger) *Handler {
	return &Handler{
		sender: sender,
		log:    log,
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

	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	actorID := middleware.ActorIDFromContext(r.Context())
	if actorID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ch := h.sender.AddClient(r.RemoteAddr, actorID)

	disconnected := r.Context().Done()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-disconnected:
			h.sender.RemoveClient(r.RemoteAddr, actorID)
			return

		case msg, ok := <-ch:
			if !ok {
				return
			}

			_, err := w.Write(msg)
			if err != nil {
				h.log.Error("sse write failed:", "err", err)
				return
			}
			flusher.Flush()

		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": heartbeat\n\n")
			if err != nil {
				h.log.Error("heartbeat write error", "err", err)
				return
			}

			flusher.Flush()
		}
	}
}
