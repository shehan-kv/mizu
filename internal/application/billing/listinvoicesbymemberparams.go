package billing

type ListInvoicesByMemberParams struct {
	ActorID   string
	MemberID  string
	Keyword   *string
	IsInvoice *bool
	Status    *string
	Limit     int
	Offset    int
}
