package invoice

import "github.com/cockroachdb/apd/v3"

type InvoiceItemResponse struct {
	Id           int64        `json:"id"`
	Description  string       `json:"description"`
	Qty          *apd.Decimal `json:"qty"`
	UnitPrice    *apd.Decimal `json:"unitPrice"`
	UnitDiscount *apd.Decimal `json:"unitDiscount"`
	DiscountType string       `json:"discountType"`
	UnitTax      *apd.Decimal `json:"unitTax"`
	TaxType      string       `json:"taxType"`
	Tax          *apd.Decimal `json:"tax"`
	Discount     *apd.Decimal `json:"discount"`
	Total        *apd.Decimal `json:"total"`
}
