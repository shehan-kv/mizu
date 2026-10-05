package billing

type BillingSummaryMetricDTO struct {
	CurrencyCode string
	Amount       string
	Count        int64
}
