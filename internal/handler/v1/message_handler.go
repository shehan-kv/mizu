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
	msgSrv *service.MessageService
}

// Creates a new instance of MessageHandler
//
// Parameters:
//   - msgSrv: a pointer to a MessageService
//
// Returns:
//   - a pointer to a new MessageHandler
func NewMessageHandler(msgSrv *service.MessageService) *MessageHandler {
	return &MessageHandler{msgSrv: msgSrv}
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
func (msgHndl *MessageHandler) GetMux(lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /{channelId}", mwChain.Handle(msgHndl.CreateMessage))

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
func (msgHndl *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {

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

	resp, err := msgHndl.msgSrv.CreateMessage(r.Context(), parsedChId, &createRequest)
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
