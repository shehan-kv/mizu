package iam

import "context"

type CredentialRepository interface {
	Add(ctx context.Context, c *Credential) error

	GetByUser(ctx context.Context, uID UserID) (*Credential, error)

	Save(ctx context.Context, c *Credential) error
}
