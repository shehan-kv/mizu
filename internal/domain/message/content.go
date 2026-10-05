package message

import "strings"

const ContentMaxLength = 4000

type Content struct {
	value string
}

func NewContent(content string) (Content, error) {
	if len(strings.TrimSpace(content)) == 0 {
		return Content{}, ErrMessageContentCannotBeEmpty
	}

	if len(content) > ContentMaxLength {
		return Content{}, ErrMessageContentTooLong
	}

	return Content{
		value: content,
	}, nil
}

func (c Content) String() string {
	return c.value
}
