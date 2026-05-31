package billing

type ListInvoicesByProjectParams struct {
	ActorID   string
	ProjectID string
	Keyword   *string
	IsInvoice *bool
	Status    *string
	Limit     int
	Offset    int
}
