package contract

import "time"

type SignatoryResponse struct {
	ID        string    `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Title     *string   `json:"title"`
	Role      string    `json:"role"`
	Image     *string   `json:"image"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}
