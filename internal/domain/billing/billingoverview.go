package billing

type BillingOverview struct {
	invoicesPaid      []BillingOverviewMetric
	invoicesPending   []BillingOverviewMetric
	invoicesAccepted  []BillingOverviewMetric
	invoicesRejected  []BillingOverviewMetric
	invoicesCancelled []BillingOverviewMetric
	quotesPending     []BillingOverviewMetric
	quotesRejected    []BillingOverviewMetric
}

func NewBillingOverview(
	invPaid []BillingOverviewMetric,
	invPending []BillingOverviewMetric,
	invAccepted []BillingOverviewMetric,
	invRejected []BillingOverviewMetric,
	invCancelled []BillingOverviewMetric,
	qPending []BillingOverviewMetric,
	qRejected []BillingOverviewMetric,
) BillingOverview {

	return BillingOverview{
		invoicesPaid:      invPaid,
		invoicesPending:   invPending,
		invoicesAccepted:  invAccepted,
		invoicesRejected:  invRejected,
		invoicesCancelled: invCancelled,
		quotesPending:     qPending,
		quotesRejected:    qRejected,
	}
}

func (b BillingOverview) InvoicesPaid() []BillingOverviewMetric {
	return b.invoicesPaid
}

func (b BillingOverview) InvoicesPending() []BillingOverviewMetric {
	return b.invoicesPending
}

func (b BillingOverview) InvoicesAccepted() []BillingOverviewMetric {
	return b.invoicesAccepted
}

func (b BillingOverview) InvoicesRejected() []BillingOverviewMetric {
	return b.invoicesRejected
}

func (b BillingOverview) InvoicesCancelled() []BillingOverviewMetric {
	return b.invoicesCancelled
}

func (b BillingOverview) QuotesPending() []BillingOverviewMetric {
	return b.quotesPending
}

func (b BillingOverview) QuotesRejected() []BillingOverviewMetric {
	return b.quotesRejected
}
