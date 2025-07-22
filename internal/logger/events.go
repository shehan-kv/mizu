package logger

// Events for structured logging
var (
	EventAuthSigninAttempt        = "auth_signin_attempt"
	EventAuthInvalidCredentials   = "auth_invalid_credentials"
	EventAuthAccountDisabled      = "auth_account_disabled"
	EventAuthUserNotFound         = "auth_user_not_found"
	EventAuthRoleNotFound         = "auth_role_not_found"
	EventAuthInternalError        = "auth_internal_error"
	EventAuthPasswordNotFound     = "auth_password_not_found"
	EventAuthLastSigninNotUpdated = "auth_last_signin_not_updated"
	EventAuthUserAuthenticated    = "auth_user_authenticated"

	EventSessionIdCreateFailed = "session_id_create_failed"
	EventSessionSetFailed      = "session_set_failed"
	EventSessionNotFound       = "session_not_found"
	EventSessionRevokeFailed   = "session_revoke_failed"

	EventCorrelationIdCreateFailed = "correlation_id_create_failed"
)
