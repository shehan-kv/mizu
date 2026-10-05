package billing

type BillingSummaryMetricResponse struct {
	CurrencyCode string `json:"currencyCode"`
	Amount       string `json:"amount"`
	Count        int64  `json:"count"`
}
