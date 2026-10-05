package billing

type BillingSummaryResponse struct {
	InvoicesPaid      []BillingSummaryMetricResponse `json:"invoicesPaid"`
	InvoicesPending   []BillingSummaryMetricResponse `json:"invoicesPending"`
	InvoicesAccepted  []BillingSummaryMetricResponse `json:"invoicesAccepted"`
	InvoicesRejected  []BillingSummaryMetricResponse `json:"invoicesRejected"`
	InvoicesCancelled []BillingSummaryMetricResponse `json:"invoicesCancelled"`
	QuotesPending     []BillingSummaryMetricResponse `json:"quotesPending"`
	QuotesRejected    []BillingSummaryMetricResponse `json:"quotesRejected"`
}
