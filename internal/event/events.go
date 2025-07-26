package event

type Event string

// Auth related events
const (
	EventSigninAttempt        Event = "signin_attempt"
	EventInvalidCredentials   Event = "invalid_credentials"
	EventAccountDisabled      Event = "account_disabled"
	EventUserAuthenticated    Event = "user_authenticated"
	EventUserNotAuthenticated Event = "user_not_authenticated"
)

// System events
const (
	EventInternalError Event = "internal_error"
)

// Resource events
const (
	EventAlreadyExists Event = "already_exists"
	EventCreateFailed  Event = "create_failed"
	EventCreateSuccess Event = "create_success"
	EventNotFound      Event = "not_found"
)
