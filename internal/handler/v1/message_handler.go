package v1

import (
	"mizu/internal/db/store"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
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

	return mux
}
