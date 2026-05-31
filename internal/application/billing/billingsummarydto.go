package billing

type BillingSummaryDTO struct {
	InvoicesPaid      []BillingSummaryMetricDTO
	InvoicesPending   []BillingSummaryMetricDTO
	InvoicesAccepted  []BillingSummaryMetricDTO
	InvoicesRejected  []BillingSummaryMetricDTO
	InvoicesCancelled []BillingSummaryMetricDTO
	QuotesPending     []BillingSummaryMetricDTO
	QuotesRejected    []BillingSummaryMetricDTO
}
