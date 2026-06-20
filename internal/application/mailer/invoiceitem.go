package mailer

type InvoiceItem struct {
	Description string

	Qty       string
	UnitPrice string

	DiscountRate string
	DiscountType string

	TaxRate string
	TaxType string

	LineGross    string
	LineDiscount string
	LineNet      string
	LineTax      string
	LineTotal    string
}
