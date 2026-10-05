package billing

type BillingOverviewMetric struct {
	currencyCode CurrencyCode
	amount       Decimal
	count        int64
}

func NewBillingOverviewMetric(currencyCode CurrencyCode, amount Decimal, count int64) BillingOverviewMetric {

	return BillingOverviewMetric{
		currencyCode: currencyCode,
		amount:       amount,
		count:        count,
	}
}

func (o BillingOverviewMetric) CurrencyCode() CurrencyCode {
	return o.currencyCode
}

func (o BillingOverviewMetric) Amount() Decimal {
	return o.amount
}

func (o BillingOverviewMetric) Count() int64 {
	return o.count
}
