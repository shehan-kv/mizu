package billing

type InvoiceID string

func NewInvoiceID(id string) (InvoiceID, error) {
	if id == "" {
		return "", ErrBillingInvoiceIDCannotBeEmpty
	}

	return InvoiceID(id), nil
}

func (i InvoiceID) String() string {
	return string(i)
}
