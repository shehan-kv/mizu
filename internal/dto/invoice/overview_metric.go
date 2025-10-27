package invoice

import "github.com/cockroachdb/apd/v3"

type OverviewMetric struct {
	CurrencyCode string       `json:"currencyCode"`
	Count        int64        `json:"count"`
	Amount       *apd.Decimal `json:"amount"`
}
