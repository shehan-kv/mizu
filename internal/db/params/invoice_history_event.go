package params

type InvoiceHistoryEvent = string

const (
	InvoiceHistoryEventCreated       InvoiceHistoryEvent = "created"
	InvoiceHistoryEventStatusChanged InvoiceHistoryEvent = "status_changed"
	InvoiceHistoryEventConverted     InvoiceHistoryEvent = "converted"
	InvoiceHistoryEventAccepted      InvoiceHistoryEvent = "accepted"
	InvoiceHistoryEventRejected      InvoiceHistoryEvent = "rejected"
	InvoiceHistoryEventCancelled     InvoiceHistoryEvent = "cancelled"
	InvoiceHistoryEventPaid          InvoiceHistoryEvent = "paid"
	InvoiceHistoryEventEmailed       InvoiceHistoryEvent = "emailed"
	InvoiceHistoryEventDownloaded    InvoiceHistoryEvent = "downloaded"
)
