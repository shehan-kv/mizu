package common

type EventType string

func (e EventType) String() string {
	return string(e)
}

type EventScope string

var (
	EventScopeInternal EventScope = "internal"
	EventScopeExternal EventScope = "external"
)

type Event interface {
	EventType() EventType
	EventScope() EventScope
}
