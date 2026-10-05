package billing

import (
	"encoding/json"
	"errors"

	"mizu/internal/application/billing"
	"mizu/internal/application/logger"
	domainbilling "mizu/internal/domain/billing"
	domainproject "mizu/internal/domain/project"
	"mizu/internal/presentation/http/rest/middleware"
	"mizu/internal/presentation/http/rest/page"
	"mizu/internal/presentation/http/rest/query"
	"mizu/internal/presentation/http/rest/response"
	"net/http"
)

type BillingHandler struct {
	billingSrv *billing.Service
	log        logger.Logger
}

func NewBillingHandler(billingSrv *billing.Service, log logger.Logger) *BillingHandler {
	return &BillingHandler{billingSrv: billingSrv, log: log}
}

func (h *BillingHandler) NewMux(authMiddleware func(http.Handler) http.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /invoices", authMiddleware(http.HandlerFunc(h.ListInvoices)))
	mux.Handle("POST /invoices/{projectID}", authMiddleware(http.HandlerFunc(h.CreateInvoice)))
	mux.Handle("GET /invoices/project/{projectID}", authMiddleware(http.HandlerFunc(h.ListInvoicesByProject)))
	mux.Handle("GET /invoices/{invoiceID}", authMiddleware(http.HandlerFunc(h.GetInvoice)))
	mux.Handle("GET /invoices/member/{memberID}", authMiddleware(http.HandlerFunc(h.ListInvoicesByMember)))
	mux.Handle("PUT /invoices/{invoiceID}/accept", authMiddleware(http.HandlerFunc(h.AcceptInvoice)))
	mux.Handle("PUT /invoices/{invoiceID}/reject", authMiddleware(http.HandlerFunc(h.RejectInvoice)))
	mux.Handle("PUT /invoices/{invoiceID}/pay", authMiddleware(http.HandlerFunc(h.PayInvoice)))
	mux.Handle("PUT /invoices/{invoiceID}/cancel", authMiddleware(http.HandlerFunc(h.CancelInvoice)))
	mux.Handle("PUT /invoices/{invoiceID}/convert", authMiddleware(http.HandlerFunc(h.ConvertToInvoice)))
	mux.Handle("POST /invoices/{invoiceID}/email", authMiddleware(http.HandlerFunc(h.EmailInvoice)))
	mux.Handle("GET /invoices/paid-count", authMiddleware(http.HandlerFunc(h.ListPaidCount)))
	mux.Handle("GET /invoices/paid-count/project/{projectID}", authMiddleware(http.HandlerFunc(h.ListPaidCountByProject)))
	mux.Handle("GET /invoices/paid-count/member/{memberID}", authMiddleware(http.HandlerFunc(h.ListPaidCountByMember)))
	mux.Handle("GET /invoices/member/{memberID}/summary", authMiddleware(http.HandlerFunc(h.GetBillingSummaryByMember)))
	mux.Handle("GET /invoices/summary", authMiddleware(http.HandlerFunc(h.GetBillingSummary)))

	return mux
}

func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	items := make([]billing.CreateItemParams, len(req.Items))
	for i := range req.Items {
		items[i] = billing.CreateItemParams{
			Description:  req.Items[i].Description,
			Qty:          req.Items[i].Qty,
			UnitPrice:    req.Items[i].UnitPrice,
			DiscountRate: req.Items[i].DiscountRate,
			DiscountType: req.Items[i].DiscountType,
			TaxRate:      req.Items[i].TaxRate,
			TaxType:      req.Items[i].TaxType,
		}
	}

	if err := h.billingSrv.CreateInvoice(r.Context(), billing.CreateInvoiceParams{
		ActorID:      actorID,
		ProjectID:    r.PathValue("projectID"),
		CurrencyCode: req.CurrencyCode,
		Note:         req.Note,
		DueDate:      req.DueDate,
		IsInvoice:    req.IsInvoice,
		Items:        items,
	}); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *BillingHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.GetInvoice(r.Context(), actorID, r.PathValue("invoiceID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toInvoiceResponse(&result))
}

func (h *BillingHandler) ListInvoicesByProject(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.ListInvoicesByProject(r.Context(), billing.ListInvoicesByProjectParams{
		ActorID:   actorID,
		ProjectID: r.PathValue("projectID"),
		Keyword:   query.ExtractString(r, "q"),
		Status:    query.ExtractString(r, "status"),
		IsInvoice: query.ExtractBool(r, "isInvoice"),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]InvoiceOverviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toInvoiceOverviewResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[InvoiceOverviewResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *BillingHandler) ListInvoicesByMember(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.ListInvoicesByMember(r.Context(), billing.ListInvoicesByMemberParams{
		ActorID:   actorID,
		MemberID:  r.PathValue("memberID"),
		Keyword:   query.ExtractString(r, "q"),
		Status:    query.ExtractString(r, "status"),
		IsInvoice: query.ExtractBool(r, "isInvoice"),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]InvoiceOverviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toInvoiceOverviewResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[InvoiceOverviewResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	p, err := page.FromQuery(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid pagination parameters")
		return
	}

	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.ListInvoicesByMember(r.Context(), billing.ListInvoicesByMemberParams{
		ActorID:   actorID,
		MemberID:  actorID,
		Keyword:   query.ExtractString(r, "q"),
		Status:    query.ExtractString(r, "status"),
		IsInvoice: query.ExtractBool(r, "isInvoice"),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	items := make([]InvoiceOverviewResponse, len(result.Items))
	for i := range result.Items {
		items[i] = toInvoiceOverviewResponse(&result.Items[i])
	}

	response.WriteJSON(w, http.StatusOK, page.PaginatedResponse[InvoiceOverviewResponse]{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       p.Page,
		Limit:      p.Limit,
	})
}

func (h *BillingHandler) AcceptInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.billingSrv.AcceptInvoice(r.Context(), actorID, r.PathValue("invoiceID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) RejectInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.billingSrv.RejectInvoice(r.Context(), actorID, r.PathValue("invoiceID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.billingSrv.PayInvoice(r.Context(), actorID, r.PathValue("invoiceID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.billingSrv.CancelInvoice(r.Context(), actorID, r.PathValue("invoiceID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) ConvertToInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.billingSrv.ConvertToInvoice(r.Context(), actorID, r.PathValue("invoiceID")); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) EmailInvoice(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	if err := h.billingSrv.EmailInvoice(r.Context(), r.PathValue("invoiceID"), actorID, actorID); err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) ListPaidCountByProject(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.ListPaidCountByProject(r.Context(), actorID, r.PathValue("projectID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	metrics := make([]MetricResponse, len(result))
	for i := range result {
		metrics[i] = toMetricResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, metrics)
}

func (h *BillingHandler) ListPaidCountByMember(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.ListPaidCountByMember(r.Context(), actorID, r.PathValue("memberID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	metrics := make([]MetricResponse, len(result))
	for i := range result {
		metrics[i] = toMetricResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, metrics)
}

func (h *BillingHandler) ListPaidCount(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.ListPaidCountByMember(r.Context(), actorID, actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	metrics := make([]MetricResponse, len(result))
	for i := range result {
		metrics[i] = toMetricResponse(&result[i])
	}

	response.WriteJSON(w, http.StatusOK, metrics)
}

func (h *BillingHandler) GetBillingSummaryByMember(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.GetBillingSummaryByMember(r.Context(), actorID, r.PathValue("memberID"))
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toBillingSummaryResponse(&result))
}

func (h *BillingHandler) GetBillingSummary(w http.ResponseWriter, r *http.Request) {
	actorID := middleware.ActorIDFromContext(r.Context())

	result, err := h.billingSrv.GetBillingSummaryByMember(r.Context(), actorID, actorID)
	if err != nil {
		h.writeServiceError(w, r.Method, r.URL.Path, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toBillingSummaryResponse(&result))
}

func (h *BillingHandler) writeServiceError(w http.ResponseWriter, method string, path string, err error) {
	switch {

	// 404
	case errors.Is(err, domainbilling.ErrBillingInvoiceNotFound):
		response.WriteError(w, http.StatusNotFound, "invoice not found")

	case errors.Is(err, domainbilling.ErrBillingCurrencyNotFound):
		response.WriteError(w, http.StatusNotFound, "currency not found")

	case errors.Is(err, domainproject.ErrProjectNotFound):
		response.WriteError(w, http.StatusNotFound, "project not found")

	// 409
	case errors.Is(err, domainbilling.ErrBillingAcceptRequiresPending):
		response.WriteError(w, http.StatusConflict, "invoice must be pending to accept")

	case errors.Is(err, domainbilling.ErrBillingRejectRequiresPending):
		response.WriteError(w, http.StatusConflict, "invoice must be pending to reject")

	case errors.Is(err, domainbilling.ErrBillingPayRequiresAccepted):
		response.WriteError(w, http.StatusConflict, "invoice must be accepted to pay")

	case errors.Is(err, domainbilling.ErrBillingCancelInvalidStatus):
		response.WriteError(w, http.StatusConflict, "invoice cannot be cancelled in its current status")

	case errors.Is(err, domainbilling.ErrBillingConvertRequiresPending):
		response.WriteError(w, http.StatusConflict, "invoice must be pending to convert")

	case errors.Is(err, domainbilling.ErrBillingAlreadyAnInvoice):
		response.WriteError(w, http.StatusConflict, "document is already an invoice")

	case errors.Is(err, domainbilling.ErrBillingConcurrentModification):
		response.WriteError(w, http.StatusConflict, "invoice was modified by another request")

	// 400
	case errors.Is(err, domainbilling.ErrBillingInvoiceMustHaveItems):
		response.WriteError(w, http.StatusBadRequest, "invoice must have at least one item")

	case errors.Is(err, domainbilling.ErrBillingDescriptionCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "item description cannot be empty")

	case errors.Is(err, domainbilling.ErrBillingQtyMustBePositive):
		response.WriteError(w, http.StatusBadRequest, "quantity must be positive")

	case errors.Is(err, domainbilling.ErrBillingInvalidDecimal):
		response.WriteError(w, http.StatusBadRequest, "invalid decimal value")

	case errors.Is(err, domainbilling.ErrBillingInvalidDecimalPlaces):
		response.WriteError(w, http.StatusBadRequest, "invalid decimal places")

	case errors.Is(err, domainbilling.ErrBillingInvalidDiscountType):
		response.WriteError(w, http.StatusBadRequest, "invalid discount type")

	case errors.Is(err, domainbilling.ErrBillingInvalidTaxType):
		response.WriteError(w, http.StatusBadRequest, "invalid tax type")

	case errors.Is(err, domainbilling.ErrBillingInvalidStatus):
		response.WriteError(w, http.StatusBadRequest, "invalid invoice status")

	case errors.Is(err, domainbilling.ErrBillingInvalidCurrencyCode):
		response.WriteError(w, http.StatusBadRequest, "invalid currency code")

	case errors.Is(err, domainbilling.ErrBillingCurrencyCodeCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "currency code cannot be empty")

	case errors.Is(err, domainbilling.ErrBillingInvoiceIDCannotBeEmpty):
		response.WriteError(w, http.StatusBadRequest, "invoice id cannot be empty")

	// 403
	case errors.Is(err, domainproject.ErrNotProjectMember):
		response.WriteError(w, http.StatusForbidden, "not a project member")

	default:
		h.log.Error(
			"internal server error",
			"method", method,
			"path", path,
			"error", err,
		)

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
	}
}

func toInvoiceItemResponse(item *billing.InvoiceItemDTO) InvoiceItemResponse {
	return InvoiceItemResponse{
		Description:           item.Description,
		Qty:                   item.Qty,
		UnitPrice:             item.UnitPrice,
		DiscountRate:          item.DiscountRate,
		DiscountType:          item.DiscountType,
		TaxRate:               item.TaxRate,
		TaxType:               item.TaxType,
		DiscountAmountPerUnit: item.DiscountAmountPerUnit,
		TaxableBasePerUnit:    item.TaxableBasePerUnit,
		TaxAmountPerUnit:      item.TaxAmountPerUnit,
		LineGross:             item.LineGross,
		LineDiscount:          item.LineDiscount,
		LineNet:               item.LineNet,
		LineTax:               item.LineTax,
		LineTotal:             item.LineTotal,
	}
}

func toInvoiceResponse(inv *billing.InvoiceDTO) InvoiceResponse {
	items := make([]InvoiceItemResponse, len(inv.Items))
	for i := range inv.Items {
		items[i] = toInvoiceItemResponse(&inv.Items[i])
	}
	return InvoiceResponse{
		ID:            inv.ID,
		ProjectID:     inv.ProjectID,
		ProjectName:   inv.ProjectName,
		IsInvoice:     inv.IsInvoice,
		Status:        inv.Status,
		DueAt:         inv.DueAt,
		CurrencyCode:  inv.CurrencyCode,
		Note:          inv.Note,
		TotalTax:      inv.TotalTax,
		TotalDiscount: inv.TotalDiscount,
		SubTotal:      inv.SubTotal,
		Items:         items,
		CreatedAt:     inv.CreatedAt,
		UpdatedAt:     inv.UpdatedAt,
	}
}

func toInvoiceOverviewResponse(inv *billing.InvoiceOverviewDTO) InvoiceOverviewResponse {
	return InvoiceOverviewResponse{
		ID:            inv.ID,
		ProjectID:     inv.ProjectID,
		ProjectName:   inv.ProjectName,
		IsInvoice:     inv.IsInvoice,
		Status:        inv.Status,
		DueAt:         inv.DueAt,
		CurrencyCode:  inv.CurrencyCode,
		Note:          inv.Note,
		TotalTax:      inv.TotalTax,
		TotalDiscount: inv.TotalDiscount,
		SubTotal:      inv.SubTotal,
		CreatedAt:     inv.CreatedAt,
		UpdatedAt:     inv.UpdatedAt,
	}
}

func toBillingOverviewMetricResponse(m *billing.BillingSummaryMetricDTO) BillingSummaryMetricResponse {
	return BillingSummaryMetricResponse{
		CurrencyCode: m.CurrencyCode,
		Amount:       m.Amount,
		Count:        m.Count,
	}
}

func toMetricSlice(metrics []billing.BillingSummaryMetricDTO) []BillingSummaryMetricResponse {
	result := make([]BillingSummaryMetricResponse, len(metrics))
	for i := range metrics {
		result[i] = toBillingOverviewMetricResponse(&metrics[i])
	}
	return result
}

func toBillingSummaryResponse(b *billing.BillingSummaryDTO) BillingSummaryResponse {
	return BillingSummaryResponse{
		InvoicesPaid:      toMetricSlice(b.InvoicesPaid),
		InvoicesPending:   toMetricSlice(b.InvoicesPending),
		InvoicesAccepted:  toMetricSlice(b.InvoicesAccepted),
		InvoicesRejected:  toMetricSlice(b.InvoicesRejected),
		InvoicesCancelled: toMetricSlice(b.InvoicesCancelled),
		QuotesPending:     toMetricSlice(b.QuotesPending),
		QuotesRejected:    toMetricSlice(b.QuotesRejected),
	}
}

func toMetricResponse(m *billing.MetricDTO) MetricResponse {
	return MetricResponse{
		Key:   m.Key,
		Value: m.Value,
	}
}
