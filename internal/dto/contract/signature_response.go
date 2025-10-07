package contract

import "time"

type SignatureResponse struct {
	Id        int64      `json:"id"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	SignedAt  *time.Time `json:"signedAt"`
	Status    string     `json:"status"`
	Image     *string    `json:"image"`
}
