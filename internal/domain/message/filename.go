package message

type FileName struct {
	value string
}

func NewFileName(name string) (FileName, error) {

	if name == "" {
		return FileName{}, ErrFileNameCannotBeEmpty
	}

	return FileName{value: name}, nil
}

func (n FileName) String() string {
	return n.value
}
