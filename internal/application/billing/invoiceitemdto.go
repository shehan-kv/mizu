package billing

type InvoiceItemDTO struct {
	Description           string
	Qty                   string
	UnitPrice             string
	DiscountRate          string
	DiscountType          string
	TaxRate               string
	TaxType               string
	DiscountAmountPerUnit string
	TaxableBasePerUnit    string
	TaxAmountPerUnit      string
	LineGross             string
	LineDiscount          string
	LineNet               string
	LineTax               string
	LineTotal             string
}
