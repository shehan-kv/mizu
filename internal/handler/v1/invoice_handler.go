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

	// eg, /invoices/project/{projectId}
	mux.Handle("POST /project/{projectId}", mwChain.Handle(invHndl.CreateInvoice))
	mux.Handle("GET /project/{projectId}", mwChain.Handle(invHndl.GetInvoicesByProject))

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

func (invHndl *InvoiceHandler) GetInvoicesByProject(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	invType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
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

	resp, err := invHndl.invSrv.GetInvoicesByProject(r.Context(), parsedId, &dto.InvoiceSearch{
		Keyword: "",
		Status:  status,
		Type:    invType,
		Page:    page,
		Limit:   limit,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
