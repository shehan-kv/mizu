package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/message"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"strconv"
)

// Handles message-related HTTP requests.
//
// Uses a MessageService to perform
// message operations
type MessageHandler struct {
	srv *service.MessageService
}

// Creates a new instance of MessageHandler
//
// Parameters:
//   - srv: a pointer to a MessageService
//
// Returns:
//   - a pointer to a new MessageHandler
func NewMessageHandler(srv *service.MessageService) *MessageHandler {
	return &MessageHandler{srv: srv}
}

// Creates a ServeMux for message routes and middleware.
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
func (h *MessageHandler) GetMux(lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /{channelId}", mwChain.Handle(h.CreateMessage))
	mux.Handle("GET /{channelId}", mwChain.Handle(h.GetMessages))
	mux.Handle("GET /channels", mwChain.Handle(h.GetChannels))
	mux.Handle("GET /members/{channelId}", mwChain.Handle(h.GetMembersByChannel))

	return mux
}

// Handles creating new message in a channel.
//
// Method: POST
//
// Possible Response Codes:
//   - 401 Unauthorized – if user doesn't have access to the channel
//   - 400 BadRequest – Invalid input or user doesn't exist in the database
//   - 500 InternalServerError - Server error
//   - 201 OK - Created successfully
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {

	channelId := r.PathValue("channelId")
	parsedChId, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil || parsedChId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.MessageCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !createRequest.Validate() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.CreateMessage(r.Context(), parsedChId, &createRequest)
	if err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetMessages handles HTTP GET requests to retrieve a paginated list of
// messages of a specified channel.
// The channel ID is expected as a path parameter, eg: messages/{channelId}.
// If the channel ID is valid, a JSON-encoded, paginated list of messages is
// returned with HTTP 200 OK status.
//
// If the channelId is missing or invalid, HTTP 400 BadRequest is returned.
// If user doesn't have access to channel, HTTP 401 Unauthorized is returned.
// If an internal error occurs, HTTP 500 InternalServerError is returned.
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {

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
		limit = 30
	} else {
		parsedLimit, err := strconv.ParseInt(strLimit, 10, 64)
		if err != nil || parsedLimit <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	channelId := r.PathValue("channelId")
	parsedChId, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil || parsedChId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetMessages(r.Context(), parsedChId, &dto.MessageSearchQuery{
		Page:  page,
		Limit: limit,
	})

	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *MessageHandler) GetChannels(w http.ResponseWriter, r *http.Request) {

	result, err := h.srv.GetChannels(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *MessageHandler) GetMembersByChannel(w http.ResponseWriter, r *http.Request) {

	channelId := r.PathValue("channelId")
	parsedChId, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil || parsedChId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.srv.GetMembersByChannel(r.Context(), parsedChId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
