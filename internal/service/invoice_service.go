package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
	"mizu/internal/dto/common"
	dto "mizu/internal/dto/invoice"
	"mizu/internal/event"
	"mizu/internal/logger"
	"mizu/internal/middleware"

	"github.com/cockroachdb/apd/v3"
)

// Handles invoice related operations.
type InvoiceService struct {
	lg    logger.Logger
	invSt store.InvoiceStore
}

// Creates a new instance of InvoiceService.
// It takes a logger, a InvoiceStore for invoice-related data operations.
//
// Parameters:
//   - lg: logger that implements the logger.Logger interface
//   - invSt: invoice store that implements the InvoiceStore interface
//
// Returns:
//   - a pointer to a new InvoiceService
func NewInvoiceService(lg logger.Logger, invSt store.InvoiceStore) *InvoiceService {
	return &InvoiceService{
		lg:    lg,
		invSt: invSt,
	}
}

// Creates a new invoice with invoice items.
//
// Parameters:
//   - ctx: context for request scoping and cancellation.
//   - request: a pointer to InvoiceCreateRequest DTO.
//
// Returns:
//   - ErrBadRequest: if database constraint violations occur.
//   - ErrInternalError: if internal errors occur.
func (invSrv *InvoiceService) CreateInvoice(ctx context.Context, projectId int64, request *dto.InvoiceCreateRequest) error {

	cid := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", cid,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	sumOfTax := apd.New(0, 0)
	sumOfDiscount := apd.New(0, 0)
	subTotal := apd.New(0, 0)

	apdCtx := apd.BaseContext.WithPrecision(19)

	var invoiceItems []params.InvoiceItem

	for _, item := range request.Items {

		itemDiscount := apd.New(0, 0)
		if item.DiscountType == "percentage" {

			discountRate := apd.New(0, 0)
			if _, err := apdCtx.Quo(discountRate, &item.Discount, apd.New(100, 0)); err != nil {

				invSrv.lg.Warn("could not calculate discount rate from percentage",
					"event", event.EventInternalError,
					"correlation_id", cid,
					"project_id", projectId,
					"item_description", item.Description,
					"scope", "invoice_service",
					"err", err)
				return ErrInternalError
			}

			if _, err := apdCtx.Mul(itemDiscount, &item.UnitPrice, discountRate); err != nil {

				invSrv.lg.Warn("could not calculate discount for item",
					"event", event.EventInternalError,
					"correlation_id", cid,
					"project_id", projectId,
					"item_description", item.Description,
					"scope", "invoice_service",
					"err", err)
				return ErrInternalError
			}
		} else {
			itemDiscount = &item.Discount
		}

		discountedPrice := apd.New(0, 0)
		if _, err := apdCtx.Sub(discountedPrice, &item.UnitPrice, itemDiscount); err != nil {

			invSrv.lg.Warn("could not calculate price after discount",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		itemTax := apd.New(0, 0)
		if item.TaxType == "percentage" {
			taxRate := apd.New(0, 0)
			if _, err := apdCtx.Quo(taxRate, &item.Tax, apd.New(100, 0)); err != nil {

				invSrv.lg.Warn("could not calculate tax rate from percentage",
					"event", event.EventInternalError,
					"correlation_id", cid,
					"project_id", projectId,
					"item_description", item.Description,
					"scope", "invoice_service",
					"err", err)
				return ErrInternalError
			}
			if _, err := apdCtx.Mul(itemTax, discountedPrice, taxRate); err != nil {

				invSrv.lg.Warn("could not calculate tax for item",
					"event", event.EventInternalError,
					"correlation_id", cid,
					"project_id", projectId,
					"item_description", item.Description,
					"scope", "invoice_service",
					"err", err)
				return ErrInternalError
			}
		} else {
			itemTax = &item.Tax
		}

		priceAfterTax := apd.New(0, 0)
		if _, err := apdCtx.Add(priceAfterTax, discountedPrice, itemTax); err != nil {

			invSrv.lg.Warn("could not calculate price after tax",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		var invoiceItem params.InvoiceItem
		invoiceItem.Description = item.Description
		invoiceItem.Qty = &item.Qty
		invoiceItem.UnitPrice = &item.UnitPrice
		invoiceItem.UnitDiscount = &item.Discount
		invoiceItem.DiscountType = item.DiscountType
		invoiceItem.UnitTax = &item.Tax
		invoiceItem.TaxType = item.TaxType
		invoiceItem.Tax = apd.New(0, 0)
		invoiceItem.Discount = apd.New(0, 0)
		invoiceItem.Total = apd.New(0, 0)

		if _, err := apdCtx.Mul(invoiceItem.Total, priceAfterTax, &item.Qty); err != nil {
			invSrv.lg.Warn("could not calculate price after tax for all qty",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		if _, err := apdCtx.Mul(invoiceItem.Tax, itemTax, &item.Qty); err != nil {
			invSrv.lg.Warn("could not calculate for all qty",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		if _, err := apdCtx.Mul(invoiceItem.Discount, itemDiscount, &item.Qty); err != nil {
			invSrv.lg.Warn("could not calculate discount for all qty",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		if _, err := apdCtx.Add(sumOfTax, sumOfTax, invoiceItem.Tax); err != nil {
			invSrv.lg.Warn("could not add tax for all qty to invoice sum of tax",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		if _, err := apdCtx.Add(sumOfDiscount, sumOfDiscount, invoiceItem.Discount); err != nil {
			invSrv.lg.Warn("could not add discount for all qty to invoice sum of discount",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		if _, err := apdCtx.Add(subTotal, subTotal, invoiceItem.Total); err != nil {
			invSrv.lg.Warn("could not add sub-total for all qty to invoice sub-total",
				"event", event.EventInternalError,
				"correlation_id", cid,
				"project_id", projectId,
				"item_description", item.Description,
				"scope", "invoice_service",
				"err", err)
			return ErrInternalError
		}

		invoiceItems = append(invoiceItems, invoiceItem)
	}

	invoice := params.InvoiceCreate{
		ProjectId:    projectId,
		IsInvoice:    request.IsInvoice,
		Status:       request.Status,
		CurrencyCode: request.CurrencyCode,
		Note:         request.Note,
		Discount:     sumOfDiscount,
		Tax:          sumOfTax,
		Total:        subTotal,
		Items:        invoiceItems,
	}

	invoiceId, err := invSrv.invSt.CreateOne(ctx, actor.Id, &invoice)
	if err != nil {
		if errors.Is(err, store.ErrNotNullViolation) {
			invSrv.lg.Warn("required field is null",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"project_id", projectId,
				"scope", "invoice_service",
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrForeignKeyViolation) {
			invSrv.lg.Warn("invoice foreign key constraint violated",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"project_id", projectId,
				"scope", "invoice_service",
				"err", err)
			return ErrBadRequest
		}

		if errors.Is(err, store.ErrCheckViolation) {
			invSrv.lg.Warn("check constraint violation",
				"event", event.EventCreateFailed,
				"correlation_id", cid,
				"project_id", projectId,
				"scope", "invoice_service",
				"err", err)
			return ErrBadRequest
		}

		invSrv.lg.Warn("could not create invoice",
			"event", event.EventCreateFailed,
			"correlation_id", cid,
			"project_id", projectId,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	invSrv.lg.Info("invoice created successfully",
		"event", event.EventCreateSuccess,
		"correlation_id", cid,
		"project_id", projectId,
		"invoice_id", invoiceId,
		"scope", "invoice_service")

	return nil
}

// GetInvoicesByProject retrieves a paginated list of invoices
// with status for the specified project. The project is specified by the ID.
// It supports type and status filtering,
// and returns results wrapped in a common.Page payload.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) GetInvoicesByProject(
	ctx context.Context,
	projectId int64,
	query *dto.InvoiceSearch) (*common.Page[[]dto.InvoiceSummaryResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"project_id", projectId,
			"err", err)
		return nil, ErrInternalError
	}

	invoiceSearch := params.InvoiceSearch{
		Keyword: "",
		Type:    query.Type,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	}

	invoices, err := invSrv.invSt.GetSummaryByProjectId(ctx, projectId, &invoiceSearch)
	if err != nil {
		invSrv.lg.Error("could not get invoice list",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	count, err := invSrv.invSt.CountSummaryByProjectId(ctx, projectId, &invoiceSearch)
	if err != nil {
		invSrv.lg.Error("could not get invoice count",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	respInv := make([]dto.InvoiceSummaryResponse, len(invoices))

	for i, invoice := range invoices {
		respInv[i] = dto.InvoiceSummaryResponse{
			Id:           invoice.Id,
			ProjectId:    invoice.ProjectId,
			ProjectName:  invoice.ProjectName,
			IsInvoice:    invoice.IsInvoice,
			IssuedAt:     invoice.IssuedAt,
			DueAt:        invoice.DueAt,
			Discount:     invoice.Discount,
			Tax:          invoice.Tax,
			Total:        invoice.Total,
			CurrencyCode: invoice.CurrencyCode,
			Status:       invoice.Status,
			Note:         invoice.Note,
		}
	}

	resp := common.Page[[]dto.InvoiceSummaryResponse]{
		Count: count,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  respInv,
	}

	return &resp, nil
}

// GetAllByUser retrieves a paginated list of invoices
// with project information of all projects assigned to the current user.
// Current user is fetched from context. This function expects middleware to
// properly add the requesting user to the request context.
// It supports typen keyword and status filtering,
// and returns results wrapped in a common.Page payload.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) GetAllByUser(
	ctx context.Context,
	query *dto.InvoiceSearch) (*common.Page[[]dto.InvoiceSummaryResponse], error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return nil, ErrInternalError
	}

	invoiceSearch := params.InvoiceSearch{
		Keyword: query.Keyword,
		Type:    query.Type,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	}

	invoices, err := invSrv.invSt.GetSummaryByUserId(ctx, actor.Id, &invoiceSearch)
	if err != nil {
		invSrv.lg.Error("could not get invoice list",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	count, err := invSrv.invSt.CountSummaryByUserId(ctx, actor.Id, &invoiceSearch)
	if err != nil {
		invSrv.lg.Error("could not get invoice count",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"status", query.Status,
			"page", query.Page,
			"limit", query.Limit,
			"err", err)
		return nil, ErrInternalError
	}

	respInv := make([]dto.InvoiceSummaryResponse, len(invoices))

	for i, invoice := range invoices {
		respInv[i] = dto.InvoiceSummaryResponse{
			Id:           invoice.Id,
			ProjectId:    invoice.ProjectId,
			ProjectName:  invoice.ProjectName,
			IsInvoice:    invoice.IsInvoice,
			IssuedAt:     invoice.IssuedAt,
			DueAt:        invoice.DueAt,
			Discount:     invoice.Discount,
			Tax:          invoice.Tax,
			Total:        invoice.Total,
			CurrencyCode: invoice.CurrencyCode,
			Status:       invoice.Status,
			Note:         invoice.Note,
		}
	}

	resp := common.Page[[]dto.InvoiceSummaryResponse]{
		Count: count,
		Limit: query.Limit,
		Page:  query.Page,
		Data:  respInv,
	}

	return &resp, nil
}

// GetOneById retrieves a detailed invoice with
// invoice items. The invoice is specified by the ID.
// Returns a pointer to a dto.InvoiceDetailsResponse.
// This method expects middleware to properly authorize requests.
//
//   - If an error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) GetOneById(
	ctx context.Context,
	invoiceId int64) (*dto.InvoiceDetailsResponse, error) {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return nil, ErrInternalError
	}

	summary, err := invSrv.invSt.GetSummaryById(ctx, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not get invoice summary",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return nil, ErrInternalError
	}

	items, err := invSrv.invSt.GetItemsByInvoiceId(ctx, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not get invoice items",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return nil, ErrInternalError
	}

	history, err := invSrv.invSt.GetHistoryByInvoiceId(ctx, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not get invoice history",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return nil, ErrInternalError
	}

	resp := dto.InvoiceDetailsResponse{
		Id:           summary.Id,
		ProjectId:    summary.ProjectId,
		ProjectName:  summary.ProjectName,
		IsInvoice:    summary.IsInvoice,
		Status:       summary.Status,
		IssuedAt:     summary.IssuedAt,
		DueAt:        summary.DueAt,
		Total:        summary.Total,
		Discount:     summary.Discount,
		Tax:          summary.Tax,
		CurrencyCode: summary.CurrencyCode,
		Note:         summary.Note,
		Items:        make([]dto.InvoiceItemResponse, len(items)),
		History:      make([]dto.InvoiceHistoryResponse, len(history)),
	}

	for i, item := range items {
		resp.Items[i] = dto.InvoiceItemResponse{
			Id:            item.Id,
			Description:   item.Description,
			Qty:           item.Qty,
			UnitPrice:     item.UnitPrice,
			UnitDiscount:  item.UnitDiscount,
			DiscountType:  item.DiscountType,
			UnitTax:       item.UnitTax,
			TaxType:       item.TaxType,
			TotalTax:      item.TotalTax,
			TotalDiscount: item.TotalDiscount,
			Total:         item.Total,
		}
	}

	for i, entry := range history {
		resp.History[i] = dto.InvoiceHistoryResponse{
			Id: entry.Id,
			User: dto.InvoiceHistoryUser{
				Id:        entry.UserId,
				FirstName: entry.FirstName,
				LastName:  entry.LastName,
				Title:     entry.Title,
				Image:     entry.Image,
				Role:      entry.Role,
			},
			Event:      entry.Event,
			RecoredAt:  entry.RecordedAt,
			IsInvoice:  entry.IsInvoice,
			LastStatus: entry.LastStatus,
			NewStatus:  entry.NewStatus,
		}
	}

	return &resp, nil
}

// AcceptById marks an invoice as accepted.
// The invoice is specified by the ID.
// The requesting user is retrieved from the context.
// This method expects middleware to properly authorize requests
// and to properly add the requesting user to the context.
//
//   - If the invoice is already accepted, it returns service.ErrAlreadyExists
//   - If any other error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) AcceptById(ctx context.Context, invoiceId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	alreadyAccepted, err := invSrv.invSt.AcceptById(ctx, actor.Id, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not accept invoice/quote",
			"event", event.EventCreateFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return ErrInternalError
	}

	if alreadyAccepted {
		return ErrAlreadyExists
	}

	return nil
}

// RejectById marks an invoice as rejected.
// The invoice is specified by the ID.
// The requesting user is retrieved from the context.
// This method expects middleware to properly authorize requests
// and to properly add the requesting user to the context.
//
//   - If the invoice is already rejected, it returns service.ErrAlreadyExists
//   - If any other error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) RejectById(ctx context.Context, invoiceId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	alreadyAccepted, err := invSrv.invSt.RejectById(ctx, actor.Id, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not reject invoice/quote",
			"event", event.EventCreateFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return ErrInternalError
	}

	if alreadyAccepted {
		return ErrAlreadyExists
	}

	return nil
}

// CancelById marks an invoice or a quote as cancelled.
// The invoice is specified by the ID.
// The requesting user is retrieved from the context.
// This method expects middleware to properly authorize requests
// and to properly add the requesting user to the context.
//
//   - If the invoice is already rejected, it returns service.ErrAlreadyExists
//   - If any other error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) CancelById(ctx context.Context, invoiceId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	alreadyAccepted, err := invSrv.invSt.CancelById(ctx, actor.Id, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not cancel invoice/quote",
			"event", event.EventCreateFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return ErrInternalError
	}

	if alreadyAccepted {
		return ErrAlreadyExists
	}

	return nil
}

// PayById marks an invoice as paid.
// The invoice is specified by the ID.
// The requesting user is retrieved from the context.
// This method expects middleware to properly authorize requests
// and to properly add the requesting user to the context.
//
//   - If the invoice is already rejected, it returns service.ErrAlreadyExists
//   - If the invoice ID points to a quote, it returns service.ErrBadRequest
//   - If any other error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) PayById(ctx context.Context, invoiceId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	alreadyPaid, err := invSrv.invSt.PayById(ctx, actor.Id, invoiceId)
	if err != nil {

		if errors.Is(err, store.ErrUnexpectedType) {
			invSrv.lg.Error("cannot pay a quote, must be an invoice",
				"event", event.EventCreateFailed,
				"scope", "invoice_service",
				"correlation_id", correlationId,
				"actor_id", actor.Id,
				"invoice_id", invoiceId,
				"err", err)
			return ErrBadRequest
		}

		invSrv.lg.Error("could not pay invoice/quote",
			"event", event.EventCreateFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return ErrInternalError
	}

	if alreadyPaid {
		return ErrAlreadyExists
	}

	return nil
}

// QuoteToInvoice converts a quote to an invoice.
// The quote is specified by the ID.
// The requesting user is retrieved from the context.
// This method expects middleware to properly authorize requests
// and to properly add the requesting user to the context.
//
//   - If the quote is already converted, it returns service.ErrAlreadyExists
//   - If the quote is in an invalid state, it returns service.ErrBadRequest
//   - If any other error occurs, it returns service.ErrInternalError
func (invSrv *InvoiceService) QuoteToInvoice(ctx context.Context, quoteId int64) error {

	correlationId := middleware.GetCorrelationID(ctx)
	actor, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		invSrv.lg.Error("could not get actor from context",
			"event", event.EventInternalError,
			"correlation_id", correlationId,
			"scope", "invoice_service",
			"err", err)
		return ErrInternalError
	}

	alreadyConverted, err := invSrv.invSt.QuoteToInvoice(ctx, actor.Id, quoteId)

	if err != nil {
		if errors.Is(err, store.ErrUnexpectedType) {
			invSrv.lg.Error("quote is in an invalid state to convert",
				"event", event.EventCreateFailed,
				"scope", "invoice_service",
				"correlation_id", correlationId,
				"actor_id", actor.Id,
				"quote_id", quoteId,
				"err", err)
			return ErrBadRequest
		}

		invSrv.lg.Error("could not convert quote to invoice",
			"event", event.EventCreateFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"quote_id", quoteId,
			"err", err)
		return ErrInternalError
	}

	if alreadyConverted {
		return ErrAlreadyExists
	}

	return nil
}
