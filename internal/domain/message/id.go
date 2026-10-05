package message

type ChannelID string

func NewChannelID(id string) (ChannelID, error) {
	if id == "" {
		return "", ErrChannelIDCannotBeEmpty
	}

	return ChannelID(id), nil
}

func (c ChannelID) String() string {
	return string(c)
}

type MessageID string

func NewMessageID(id string) (MessageID, error) {
	if id == "" {
		return "", ErrMessageIDCannotBeEmpty
	}

	return MessageID(id), nil
}

func (m MessageID) String() string {
	return string(m)
}

type FileID string

func NewFileID(id string) (FileID, error) {
	if id == "" {
		return "", ErrFileIDCannotBeEmpty
	}

	return FileID(id), nil
}

func (m FileID) String() string {
	return string(m)
}
