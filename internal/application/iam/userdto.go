package iam

import "time"

type UserDTO struct {
	ID         string
	FirstName  string
	LastName   string
	Email      string
	Title      *string
	Role       string
	HasImage   bool
	IsActive   bool
	IsVerified bool
	CreatedAt  time.Time
	LastSignIn *time.Time
}
