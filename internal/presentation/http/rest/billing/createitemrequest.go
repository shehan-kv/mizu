package billing

type CreateItemRequest struct {
	Description  string `json:"description"`
	Qty          string `json:"qty"`
	UnitPrice    string `json:"unitPrice"`
	DiscountRate string `json:"discountRate"`
	DiscountType string `json:"discountType"`
	TaxRate      string `json:"taxRate"`
	TaxType      string `json:"taxType"`
}
