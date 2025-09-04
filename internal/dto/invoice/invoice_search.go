package invoice

type InvoiceSearch struct {
	Keyword string
	Type    string
	Status  string
	Page    int64
	Limit   int64
}
