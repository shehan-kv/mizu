package service

import (
	"context"
	"errors"
	"mizu/internal/db/params"
	"mizu/internal/db/store"
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
