package iam

import (
	"mizu/internal/domain/common"
	"time"
)

type User struct {
	id           UserID
	name         Name
	email        Email
	title        *string
	role         Role
	image        *Image
	isActive     bool
	isVerified   bool
	lastSignInAt *time.Time
	events       []common.Event

	version   int
	createdAt time.Time
	updatedAt time.Time
}

func NewSystemUser(
	id UserID,
	name Name,
	email Email,
	title *string,
	role Role,
	isActive bool,
	now time.Time) *User {

	return &User{
		id:           id,
		name:         name,
		email:        email,
		title:        title,
		role:         role,
		image:        nil,
		isActive:     isActive,
		isVerified:   false,
		lastSignInAt: nil,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}
}

func NewUser(
	id UserID,
	name Name,
	email Email,
	title *string,
	role Role,
	isActive bool,
	actor UserID,
	now time.Time) *User {

	u := User{
		id:           id,
		name:         name,
		email:        email,
		title:        title,
		role:         role,
		image:        nil,
		isActive:     isActive,
		isVerified:   false,
		lastSignInAt: nil,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}

	u.events = append(u.events, UserCreatedEvent{
		UserID:     id,
		Name:       name,
		Email:      email,
		CreatedBy:  actor,
		OccurredAt: now,
	})

	return &u
}

func RestoreUser(
	id UserID,
	name Name,
	email Email,
	title *string,
	role Role,
	image *Image,
	isActive bool,
	isVerified bool,
	lastSignInAt *time.Time,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		name:         name,
		email:        email,
		title:        title,
		role:         role,
		image:        image,
		isActive:     isActive,
		isVerified:   isVerified,
		lastSignInAt: lastSignInAt,
		version:      version,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) FirstName() string {
	return u.name.FirstName()
}

func (u *User) LastName() string {
	return u.name.LastName()
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Title() *string {
	return u.title
}

func (u *User) Role() Role {
	return u.role
}

func (u *User) Image() *Image {
	return u.image
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) IsVerified() bool {
	return u.isVerified
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) LastSignInAt() *time.Time {
	return u.lastSignInAt
}

func (u *User) Version() int {
	return u.version
}

func (u *User) IsAdministrator() bool {
	return u.role == RoleAdministrator
}

func (u *User) IsStaff() bool {
	return u.role == RoleStaff
}

func (u *User) IsClient() bool {
	return u.role == RoleClient
}

func (u *User) Activate(now time.Time) {
	u.isActive = true
	u.updatedAt = now
}

func (u *User) Deactivate(now time.Time) {
	u.isActive = false
	u.updatedAt = now
}

func (u *User) Verify(now time.Time) {

	u.events = append(u.events, UserVerifiedEvent{
		UserID:     u.id,
		OccurredAt: now,
	})

	u.isVerified = true
	u.updatedAt = now
}

func (u *User) RecordSignIn(now time.Time) {
	u.lastSignInAt = &now
	u.updatedAt = now
}

func (u *User) ChangeName(name Name, now time.Time) {
	u.name = name
	u.updatedAt = now
}

func (u *User) AssignRole(role Role, now time.Time) {
	u.role = role
	u.updatedAt = now
}

func (u *User) ChangeEmail(email Email, now time.Time) {
	if u.email == email {
		return
	}

	u.email = email
	u.isVerified = false
	u.updatedAt = now

	u.events = append(u.events, UserEmailChangedEvent{
		UserID:     u.id,
		NewEmail:   email,
		OccurredAt: now,
	})
}

func (u *User) ChangeTitle(title *string, now time.Time) {
	u.title = title
	u.updatedAt = now
}

func (u *User) ChangeImage(image *Image, now time.Time) {
	u.image = image
	u.updatedAt = now
}

func (u *User) PullEvents() []common.Event {
	events := u.events
	u.events = nil

	return events
}

func (u *User) Equals(user *User) bool {
	return u.id == user.ID()
}
