package iam

import (
	"context"
)

type Service struct {
	pwHasher  PasswordHasher
	credsRepo CredentialRepository
}

func NewService(pwHasher PasswordHasher, credsRepo CredentialRepository) *Service {
	return &Service{pwHasher: pwHasher, credsRepo: credsRepo}
}

func (s *Service) Authenticate(ctx context.Context, u *User, p PlainPassword) error {

	creds, err := s.credsRepo.GetByUser(ctx, u.ID())
	if err != nil {
		return err
	}

	if !s.pwHasher.Verify(creds.Hash(), p) {
		return ErrUserInvalidCredentials
	}

	if !u.IsActive() {
		return ErrUserInactive
	}

	return nil
}
