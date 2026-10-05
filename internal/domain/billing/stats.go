package billing

type Stats struct {
	invoiceCount     int
	invoicePaidCount int
	quoteCount       int
}

func NewStats(invoiceCount int, invoicePaidCount int, quoteCount int) Stats {
	return Stats{
		invoiceCount:     invoiceCount,
		invoicePaidCount: invoicePaidCount,
		quoteCount:       quoteCount,
	}
}

func (s Stats) InvoiceCount() int {
	return s.invoiceCount
}

func (s Stats) InvoicePaidCount() int {
	return s.invoicePaidCount
}

func (s Stats) QuoteCount() int {
	return s.quoteCount
}
