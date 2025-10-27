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

	mux.Handle("GET /", mwChain.Handle(invHndl.GetAllByUser))
	mux.Handle("GET /{invoiceId}", mwChain.Handle(invHndl.GetOneById))
	mux.Handle("POST /accept/{invoiceId}", mwChain.Handle(invHndl.AcceptById))
	mux.Handle("POST /reject/{invoiceId}", mwChain.Handle(invHndl.RejectById))
	mux.Handle("POST /cancel/{invoiceId}", mwChain.Handle(invHndl.CancelById))
	mux.Handle("POST /pay/{invoiceId}", mwChain.Handle(invHndl.PayById))
	mux.Handle("POST /quote-to-invoice/{quoteId}", mwChain.Handle(invHndl.QuoteToInvoice))

	// eg, /invoices/project/{projectId}
	mux.Handle("GET /project/{projectId}", mwChain.Handle(invHndl.GetInvoicesByProject))
	mux.Handle("POST /project/{projectId}", mwChain.Handle(invHndl.CreateInvoice))
	mux.Handle("GET /metrics/paid/{projectId}", mwChain.Handle(invHndl.GetPaidMetricsByProjectId))
	mux.Handle("GET /metrics/paid", mwChain.Handle(invHndl.GetPaidMetricsByCurrentUser))

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
//   - 201 Created - Created successfully
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

	w.WriteHeader(http.StatusCreated)
}

// GetInvoicesByProject handles HTTP GET requests for getting a paginated
// list of invoices for a specified project.
// The project ID is expected as a path parameter, eg: {projectId}.
// Supports query parameters for pagination and
// filtering results by type and status, eg: ?type=invoice&page=1.
// This method expects middleware to properly authorize requests.
//
// Supported query parameters:
//   - type: type to filter by, could be invoice or a quote
//   - status: status of invoices to filter by
//   - page: the page number requested
//   - limit: the number of results per page
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If projectId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If page or limit query params are invalid, HTTP 400 BadRequest is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
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

// GetAllByUser handles HTTP GET requests for getting a paginated
// list of invoices of all projects assigned to the requesting user.
// Supports query parameters for pagination and
// filtering results by keyword, type and status, eg: ?type=invoice&page=1.
// This method expects middleware to properly authorize requests.
//
// Supported query parameters:
//   - q: keyword to filter results by
//   - type: type to filter by, could be invoice or a quote
//   - status: status of invoices to filter by
//   - page: the page number requested
//   - limit: the number of results per page
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If page or limit query params are invalid, HTTP 400 BadRequest is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) GetAllByUser(w http.ResponseWriter, r *http.Request) {

	keyword := r.URL.Query().Get("q")
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

	resp, err := invHndl.invSrv.GetAllByUser(r.Context(), &dto.InvoiceSearch{
		Keyword: keyword,
		Type:    invType,
		Status:  status,
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

// GetOneById handles HTTP GET requests for getting a detailed invoice
// with invoice items by invoice ID.
// The invoice ID is expected as a path parameter, eg: {invoiceId}.
// This method expects middleware to properly authorize requests.
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If invoiceId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) GetOneById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := invHndl.invSrv.GetOneById(r.Context(), parsedId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// AcceptById handles HTTP POST requests for accepting an invoice by invoice ID.
// The invoice ID is expected as a path parameter, eg: {invoiceId}.
// This method expects middleware to properly authorize requests.
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If invoiceId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If invoice is already accepted, HTTP 409 Conflict is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) AcceptById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = invHndl.invSrv.AcceptById(r.Context(), parsedId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RejectById handles HTTP POST requests for rejecting an invoice by invoice ID.
// The invoice ID is expected as a path parameter, eg: {invoiceId}.
// This method expects middleware to properly authorize requests.
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If invoiceId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If invoice is already rejected, HTTP 409 Conflict is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) RejectById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = invHndl.invSrv.RejectById(r.Context(), parsedId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CancelById handles HTTP POST requests for cancelling an invoice
// or quote by invoice ID.
// The invoice ID is expected as a path parameter, eg: {invoiceId}.
// This method expects middleware to properly authorize requests.
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If invoiceId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If invoice is already cancelled, HTTP 409 Conflict is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) CancelById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = invHndl.invSrv.CancelById(r.Context(), parsedId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PayById handles HTTP POST requests for marking an invoice as paid.
// The invoice ID is expected as a path parameter, eg: {invoiceId}.
// This method expects middleware to properly authorize requests.
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If invoiceId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If invoiceId is the ID of a quote, HTTP 400 BadRequest is returned.
//   - If invoice is already paid, HTTP 409 Conflict is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) PayById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = invHndl.invSrv.PayById(r.Context(), parsedId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// QuoteToInvoice handles HTTP POST requests for converting a quote to an invoice.
// The quote ID is expected as a path parameter, eg: {quoteId}.
// This method expects middleware to properly authorize requests.
//
// HTTP responses:
//   - If the request is successful, HTTP 200 is returned.
//   - If quoteId is missing or invalid, HTTP 400 BadRequest is returned.
//   - If quote is in an invalid state to be converted, HTTP 400 BadRequest is returned.
//   - If quote is already converted, HTTP 409 Conflict is returned.
//   - If an internal error occurs, HTTP 500 InternalServerError is returned.
func (invHndl *InvoiceHandler) QuoteToInvoice(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("quoteId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = invHndl.invSrv.QuoteToInvoice(r.Context(), parsedId)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, service.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (invHndl *InvoiceHandler) GetPaidMetricsByProjectId(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := invHndl.invSrv.GetPaidCountByProjectId(r.Context(), parsedId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (invHndl *InvoiceHandler) GetPaidMetricsByCurrentUser(w http.ResponseWriter, r *http.Request) {

	resp, err := invHndl.invSrv.GetPaidMetricsByCurrentUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
