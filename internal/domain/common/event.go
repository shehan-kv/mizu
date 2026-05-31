package common

type EventType string

func (e EventType) String() string {
	return string(e)
}

type Event interface {
	EventType() EventType
}
