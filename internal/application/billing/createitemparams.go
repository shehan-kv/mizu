package billing

type CreateItemParams struct {
	Description  string
	Qty          string
	UnitPrice    string
	DiscountRate string
	DiscountType string
	TaxRate      string
	TaxType      string
}
