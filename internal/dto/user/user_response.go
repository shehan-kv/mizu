package user

import "time"

type UserResponse struct {
	Id        int64      `json:"id"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Title     *string    `json:"title"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Image     *string    `json:"image"`
	CreatedAt time.Time  `json:"createdAt"`
	LastLogin *time.Time `json:"lastLogin"`
	IsActive  bool       `json:"isActive"`
}
