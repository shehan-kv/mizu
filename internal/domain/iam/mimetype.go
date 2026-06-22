package iam

type MimeType string

func NewMimeType(mimeType string) (MimeType, error) {
	if mimeType == "" {
		return MimeType(""), ErrUserMimeTypeCannotBeEmpty
	}

	return MimeType(mimeType), nil
}

func (n MimeType) String() string {
	return string(n)
}
