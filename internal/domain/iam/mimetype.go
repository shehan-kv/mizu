package iam

type MimeType string

func NewMimeType(mimeType string) (MimeType, error) {
	switch mimeType {
	case
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
		"image/avif":
		return MimeType(mimeType), nil
	default:
		return MimeType(""), ErrUserInvalidImageMimeType
	}

}

func (n MimeType) String() string {
	return string(n)
}
