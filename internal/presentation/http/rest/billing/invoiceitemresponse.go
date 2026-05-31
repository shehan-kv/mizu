package billing

type InvoiceItemResponse struct {
	Description           string `json:"description"`
	Qty                   string `json:"qty"`
	UnitPrice             string `json:"unitPrice"`
	DiscountRate          string `json:"discountRate"`
	DiscountType          string `json:"discountType"`
	TaxRate               string `json:"taxRate"`
	TaxType               string `json:"taxType"`
	DiscountAmountPerUnit string `json:"discountAmountPerUnit"`
	TaxableBasePerUnit    string `json:"taxableBasePerUnit"`
	TaxAmountPerUnit      string `json:"taxAmountPerUnit"`
	LineGross             string `json:"lineGross"`
	LineDiscount          string `json:"lineDiscount"`
	LineNet               string `json:"lineNet"`
	LineTax               string `json:"lineTax"`
	LineTotal             string `json:"lineTotal"`
}
