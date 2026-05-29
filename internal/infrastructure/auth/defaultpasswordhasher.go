package auth

import (
	"mizu/internal/domain/iam"

	"golang.org/x/crypto/bcrypt"
)

// Hashing algorithm prefix constants.
// String type assigned so there's no need to convert string to int for comparison
const (
	bcryptHashPrefix = "00"
)

type DefaultPasswordHasher struct {
}

func NewDefaultPasswordHasher() *DefaultPasswordHasher {
	return &DefaultPasswordHasher{}
}

// Hash calls the currently used hashing algorithm to hash the plain-text password
func (h *DefaultPasswordHasher) Hash(password iam.PlainPassword) (string, error) {
	return h.bcryptHash(password.Value())
}

// Verify compares a hashed password with a plain-text password.
// The hash string is expected to have a 2-character prefix that indicates
// the hashing algorithm used (e.g., "00" for bcrypt etc.).
// The function extracts the prefix, selects the appropriate validation function,
// and compares the provided plain-text password against the stored hash.
//
// Returns true if the password is valid, false otherwise.
func (h *DefaultPasswordHasher) Verify(hash string, password iam.PlainPassword) bool {

	switch hash[:2] {
	case bcryptHashPrefix:
		return h.bcryptValidate(hash[2:], password.Value())

	default:
		return false
	}

}

// bcryptHash hashes the given plain-text password using the Bcrypt algorithm.
// Adds a prefix to the hash so the decoding function can recognize the algorithm
// The resulting hash has the format 00<generated_hash> where 00 is the prefix
//
// Returns the prefixed hash and any error encountered during hashing.

func (h *DefaultPasswordHasher) bcryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return bcryptHashPrefix + string(bytes), err
}

// bcryptValidate validates a hashed Bcrypt password
func (h *DefaultPasswordHasher) bcryptValidate(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
