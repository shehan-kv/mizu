package invoice

type InvoiceOverviewResponse struct {
	Paid           []OverviewMetric `json:"paid"`
	Pending        []OverviewMetric `json:"pending"`
	Accepted       []OverviewMetric `json:"accepted"`
	Rejected       []OverviewMetric `json:"rejected"`
	Cancelled      []OverviewMetric `json:"cancelled"`
	QuotesRejected []OverviewMetric `json:"quotesRejected"`
	QuotesPending  []OverviewMetric `json:"quotesPending"`
}
