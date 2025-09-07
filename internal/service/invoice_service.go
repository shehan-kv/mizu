package service

import (
	"context"
	"errors"
	"math"
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

	invoiceId, err := invSrv.invSt.CreateOne(ctx, &invoice)
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
	query *dto.InvoiceSearch) (*common.Page[[]dto.InvoiceWithStatusResponse], error) {

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

	invoices, err := invSrv.invSt.GetInvoiceStatsByProject(ctx, projectId, &params.InvoiceSearch{
		Keyword: "",
		Type:    query.Type,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

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

	respInv := make([]dto.InvoiceWithStatusResponse, len(invoices.Items))

	for i, invoice := range invoices.Items {
		respInv[i] = dto.InvoiceWithStatusResponse{
			Id:           invoice.Id,
			IsInvoice:    invoice.IsInvoice,
			IssuedAt:     invoice.IssuedAt,
			DueAt:        invoice.DueAt,
			Total:        invoice.Total,
			CurrencyCode: invoice.CurrencyCode,
			Status:       invoice.Status,
		}
	}

	numOfPages := math.Ceil(float64(invoices.Total) / float64(query.Limit))
	resp := common.Page[[]dto.InvoiceWithStatusResponse]{
		CurrentPage: query.Page,
		Limit:       query.Limit,
		TotalPages:  int64(numOfPages),
		Data:        respInv,
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
	query *dto.InvoiceSearch) (*common.Page[[]dto.InvoiceWithProjectResponse], error) {

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

	invoices, err := invSrv.invSt.GetWithProjectByUserId(ctx, actor.Id, &params.InvoiceSearch{
		Keyword: query.Keyword,
		Type:    query.Type,
		Status:  query.Status,
		Offset:  (query.Page - 1) * query.Limit,
		Limit:   query.Limit,
	})

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

	respInv := make([]dto.InvoiceWithProjectResponse, len(invoices.Items))

	for i, invoice := range invoices.Items {
		respInv[i] = dto.InvoiceWithProjectResponse{
			Id:           invoice.Id,
			ProjectId:    invoice.ProjectId,
			ProjectName:  invoice.ProjectName,
			IsInvoice:    invoice.IsInvoice,
			IssuedAt:     invoice.IssuedAt,
			DueAt:        invoice.DueAt,
			Total:        invoice.Total,
			CurrencyCode: invoice.CurrencyCode,
			Status:       invoice.Status,
		}
	}

	numOfPages := math.Ceil(float64(invoices.Total) / float64(query.Limit))
	resp := common.Page[[]dto.InvoiceWithProjectResponse]{
		CurrentPage: query.Page,
		Limit:       query.Limit,
		TotalPages:  int64(numOfPages),
		Data:        respInv,
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

	result, err := invSrv.invSt.GetWithDetailsById(ctx, invoiceId)
	if err != nil {
		invSrv.lg.Error("could not get invoice details",
			"event", event.EventGetFailed,
			"scope", "invoice_service",
			"correlation_id", correlationId,
			"actor_id", actor.Id,
			"invoice_id", invoiceId,
			"err", err)
		return nil, ErrInternalError
	}

	resp := dto.InvoiceDetailsResponse{
		Id:           result.Id,
		ProjectId:    result.ProjectId,
		ProjectName:  result.ProjectName,
		IsInvoice:    result.IsInvoice,
		Status:       result.Status,
		IssuedAt:     result.IssuedAt,
		DueAt:        result.DueAt,
		Total:        result.Total,
		Discount:     result.Discount,
		Tax:          result.Tax,
		CurrencyCode: result.CurrencyCode,
		Note:         result.Note,
		Items:        make([]dto.InvoiceItemResponse, len(result.Items)),
	}

	for i, item := range result.Items {
		resp.Items[i] = dto.InvoiceItemResponse{
			Id:           item.Id,
			Description:  item.Description,
			Qty:          item.Qty,
			UnitPrice:    item.Qty,
			UnitDiscount: item.UnitDiscount,
			DiscountType: item.DiscountType,
			UnitTax:      item.UnitTax,
			TaxType:      item.TaxType,
			Tax:          item.Tax,
			Discount:     item.Discount,
			Total:        item.Total,
		}
	}

	return &resp, nil
}
