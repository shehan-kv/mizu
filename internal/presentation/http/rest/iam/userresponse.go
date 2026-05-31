package iam

import "time"

type UserResponse struct {
	ID         string    `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	Title      *string   `json:"title"`
	Role       string    `json:"role"`
	Image      *string   `json:"image"`
	IsActive   bool      `json:"isActive"`
	IsVerified bool      `json:"isVerified"`
	CreatedAt  time.Time `json:"createdAt"`
}
