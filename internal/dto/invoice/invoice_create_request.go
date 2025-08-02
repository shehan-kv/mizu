package invoice

import (
	"slices"
	"strings"

	"github.com/cockroachdb/apd/v3"
)

// Represents a project create request
type InvoiceCreateRequest struct {
	IsInvoice    bool   `json:"isInvoice"`
	Status       string `json:"status"`
	CurrencyCode string `json:"currencyCode"`
	Note         string `json:"note"`
	Items        []struct {
		Description  string      `json:"description"`
		Qty          apd.Decimal `json:"qty"`
		UnitPrice    apd.Decimal `json:"unitPrice"`
		Tax          apd.Decimal `json:"tax"`
		TaxType      string      `json:"taxType"`
		Discount     apd.Decimal `json:"discount"`
		DiscountType string      `json:"discountType"`
	} `json:"items"`
}

// Validates the request against a set of acceptable
// parameters.
//
// Returns:
//   - true if valid, false otherwise
func (r *InvoiceCreateRequest) Validate() bool {
	r.format()

	if len(r.Status) == 0 {
		return false
	}

	if len(r.CurrencyCode) == 0 {
		return false
	}

	if len(r.Items) == 0 {
		return false
	}

	if !slices.Contains(
		[]string{"paid", "pending", "accepted", "rejected", "cancelled"},
		r.Status) {

		return false
	}

	for _, item := range r.Items {
		if item.Qty.Negative || item.Tax.Negative || item.UnitPrice.Negative {
			return false
		}

		if len(item.Description) == 0 ||
			!slices.Contains([]string{"fixed", "percentage"}, item.TaxType) ||
			!slices.Contains([]string{"fixed", "percentage"}, item.DiscountType) {

			return false
		}

		if !validateDecimal(&item.Qty, 19, 4) || !validateDecimal(&item.Tax, 19, 4) ||
			!validateDecimal(&item.UnitPrice, 19, 4) || !validateDecimal(&item.Discount, 19, 4) {

			return false
		}
	}

	return true
}

func (r *InvoiceCreateRequest) format() {
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
	r.CurrencyCode = strings.ToLower(strings.TrimSpace(r.CurrencyCode))

	for _, item := range r.Items {
		item.Description = strings.TrimSpace(item.Description)
		item.DiscountType = strings.ToLower(strings.TrimSpace(item.DiscountType))
		item.TaxType = strings.ToLower(strings.TrimSpace(item.TaxType))
	}
}
