package iam

import (
	"context"
	"errors"
)

type Service struct {
	pwHasher PasswordHasher
	iamRepo  Repository
}

func NewService(pwHasher PasswordHasher, iamRepo Repository) *Service {
	return &Service{pwHasher: pwHasher, iamRepo: iamRepo}
}

func (s *Service) Authenticate(ctx context.Context, email Email, password PlainPassword) (*User, error) {

	creds, err := s.iamRepo.GetCredentialsByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserInvalidCredentials
		}
		return nil, err
	}

	if !s.pwHasher.Verify(creds.hashedPassword, password) {
		return nil, ErrUserInvalidCredentials
	}

	if !creds.user.IsActive() {
		return nil, ErrUserInactive
	}

	return creds.user, nil

}
