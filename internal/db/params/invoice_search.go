package params

type InvoiceSearch struct {
	Keyword string
	Type    string
	Status  string
	Offset  int64
	Limit   int64
}
