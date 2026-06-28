package iam

import "errors"

var (
	ErrUserFirstNameCannotBeEmpty = errors.New("user firstname cannot be empty")
	ErrUserConcurrentModification = errors.New("user concurrent modification")
	ErrUserImageNameCannotBeEmpty = errors.New("user image name cannot be empty")
	ErrUserLastNameCannotBeEmpty  = errors.New("user lastname cannot be empty")
	ErrUserInvalidEmailAddress    = errors.New("user invalid email address")
	ErrUserEmailAlreadyExists     = errors.New("user email already exists")
	ErrUserEmailCannotBeEmpty     = errors.New("user email cannot be empty")
	ErrUserInvalidCredentials     = errors.New("user invalid credentials")
	ErrUserPasswordTooShort       = errors.New("user password too short")
	ErrUserCannotDeleteSelf       = errors.New("user cannot delete self")
	ErrUserPasswordTooLong        = errors.New("user password too long")
	ErrUserInvalidImageMimeType   = errors.New("user invalid image mime type")
	ErrUserAlreadyVerified        = errors.New("user is already verified")
	ErrUserIDCannotBeEmpty        = errors.New("user id cannot be empty")
	ErrUserImageNotFound          = errors.New("user image not found")
	ErrUserInvalidRole            = errors.New("user invalid role")
	ErrUserNotFound               = errors.New("user not found")
	ErrUserInactive               = errors.New("user is inactive")

	ErrCredentialConcurrentModification = errors.New("credential concurrent modification")
	ErrCredentialNotFound               = errors.New("credential not found")
)
