package message

type ChannelName struct {
	value string
}

func NewChannelName(name string) (ChannelName, error) {

	if name == "" {
		return ChannelName{}, ErrChannelNameCannotBeEmpty
	}

	return ChannelName{value: name}, nil
}

func (n ChannelName) String() string {
	return n.value
}
