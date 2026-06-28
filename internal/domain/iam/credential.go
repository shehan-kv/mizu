package iam

type Credential struct {
	userID  UserID
	hash    string
	version int
}

func NewCredential(userID UserID, hash string) *Credential {
	return &Credential{
		userID:  userID,
		hash:    hash,
		version: 1,
	}
}

func RestoreCredential(userID UserID, hash string, version int) *Credential {
	return &Credential{
		userID:  userID,
		hash:    hash,
		version: version,
	}
}

func (c *Credential) UserID() UserID {
	return c.userID
}

func (c *Credential) Hash() string {
	return c.hash
}

func (c *Credential) UpdateHash(hash string) {
	c.hash = hash
}

func (c *Credential) Version() int {
	return c.version
}
