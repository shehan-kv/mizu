package iam

type PasswordHasher interface {
	Hash(plain PlainPassword) (string, error)
	Verify(hashed string, password PlainPassword) bool
}
