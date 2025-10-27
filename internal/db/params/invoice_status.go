package params

type InvoiceStatus = string

const (
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusPending   InvoiceStatus = "pending"
	InvoiceStatusAccepted  InvoiceStatus = "accepted"
	InvoiceStatusRejected  InvoiceStatus = "rejected"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)
