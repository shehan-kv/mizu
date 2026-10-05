package iam

type ImageName string

func NewImageName(name string) (ImageName, error) {
	if name == "" {
		return ImageName(""), ErrUserImageNameCannotBeEmpty
	}

	return ImageName(name), nil
}

func (n ImageName) String() string {
	return string(n)
}
