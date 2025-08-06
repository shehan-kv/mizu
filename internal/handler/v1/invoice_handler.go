package v1

import (
	"encoding/json"
	"errors"
	"mizu/internal/db/store"
	dto "mizu/internal/dto/invoice"
	"mizu/internal/logger"
	"mizu/internal/middleware"
	"mizu/internal/service"
	"mizu/internal/session"
	"net/http"
	"strconv"
)

// Handles invoice-related HTTP requests.
//
// Uses an InvoiceService to perform
// invoice operations
type InvoiceHandler struct {
	invSrv *service.InvoiceService
}

// Creates a new instance of InvoiceHandler
//
// Parameters:
//   - invSrv: a pointer to a InvoiceService
//
// Returns:
//   - a pointer to a new InvoiceHandler
func NewInvoiceHandler(invSrv *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		invSrv: invSrv,
	}
}

// Creates a ServeMux for the invoice routes and middleware.
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
func (invHndl *InvoiceHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("POST /project/{projectId}", mwChain.Handle(invHndl.CreateInvoice))

	return mux
}

// Handles creating an invoice.
//
// Expects a JSON body of InvoiceCreateRequest DTO.
//
// Method: POST
//
// Possible Response Codes:
//   - 400 BadRequest – Invalid input, missing fields or constraint violations
//   - 500 InternalServerError - Server error
//   - 200 OK - Created successfully
func (invHndl *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createRequest dto.InvoiceCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ok := createRequest.Validate(); !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := invHndl.invSrv.CreateInvoice(r.Context(), parsedId, &createRequest); err != nil {
		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
