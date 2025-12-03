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
	srv *service.InvoiceService
}

// Creates a new instance of InvoiceHandler
//
// Parameters:
//   - srv: a pointer to a InvoiceService
//
// Returns:
//   - a pointer to a new InvoiceHandler
func NewInvoiceHandler(srv *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{
		srv: srv,
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
func (h *InvoiceHandler) GetMux(
	lg logger.Logger,
	seSt session.SessionStore,
	usrSt store.UserStore) *http.ServeMux {

	mwChain := middleware.NewChain()
	mwChain.Add(
		middleware.CorrelationId(lg),
		middleware.Authenticated(lg, seSt, usrSt))

	mux := http.NewServeMux()

	mux.Handle("GET /", mwChain.Handle(h.GetAllByUser))
	mux.Handle("GET /{invoiceId}", mwChain.Handle(h.GetOneById))
	mux.Handle("PUT /{invoiceId}/status/accepted", mwChain.Handle(h.AcceptById))
	mux.Handle("PUT /{invoiceId}/status/rejected", mwChain.Handle(h.RejectById))
	mux.Handle("PUT /{invoiceId}/status/cancelled", mwChain.Handle(h.CancelById))
	mux.Handle("PUT /{invoiceId}/status/paid", mwChain.Handle(h.PayById))
	mux.Handle("POST /quote-to-invoice/{quoteId}", mwChain.Handle(h.QuoteToInvoice))

	// eg, /invoices/project/{projectId}
	mux.Handle("GET /project/{projectId}", mwChain.Handle(h.GetInvoicesByProject))
	mux.Handle("POST /project/{projectId}", mwChain.Handle(h.CreateInvoice))
	mux.Handle("GET /metrics/paid/{projectId}", mwChain.Handle(h.GetPaidMetricsByProjectId))
	mux.Handle("GET /metrics/paid", mwChain.Handle(h.GetPaidMetricsByCurrentUser))
	mux.Handle("GET /metrics/overview", mwChain.Handle(h.GetOverview))

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
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {

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

	if err := h.srv.CreateInvoice(r.Context(), parsedId, &createRequest); err != nil {
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
func (h *InvoiceHandler) GetInvoicesByProject(w http.ResponseWriter, r *http.Request) {

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

	resp, err := h.srv.GetInvoicesByProject(r.Context(), parsedId, &dto.InvoiceSearch{
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
func (h *InvoiceHandler) GetAllByUser(w http.ResponseWriter, r *http.Request) {

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

	resp, err := h.srv.GetAllByUser(r.Context(), &dto.InvoiceSearch{
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
func (h *InvoiceHandler) GetOneById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetOneById(r.Context(), parsedId)
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
func (h *InvoiceHandler) AcceptById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.AcceptById(r.Context(), parsedId)
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
func (h *InvoiceHandler) RejectById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.RejectById(r.Context(), parsedId)
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
func (h *InvoiceHandler) CancelById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.CancelById(r.Context(), parsedId)
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
func (h *InvoiceHandler) PayById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("invoiceId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.PayById(r.Context(), parsedId)
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
func (h *InvoiceHandler) QuoteToInvoice(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("quoteId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.srv.QuoteToInvoice(r.Context(), parsedId)
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

func (h *InvoiceHandler) GetPaidMetricsByProjectId(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("projectId")
	parsedId, err := strconv.ParseInt(id, 10, 64)
	if err != nil || parsedId < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := h.srv.GetPaidCountByProjectId(r.Context(), parsedId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *InvoiceHandler) GetPaidMetricsByCurrentUser(w http.ResponseWriter, r *http.Request) {

	resp, err := h.srv.GetPaidMetricsByCurrentUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *InvoiceHandler) GetOverview(w http.ResponseWriter, r *http.Request) {

	resp, err := h.srv.GetOverview(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
