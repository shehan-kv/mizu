package auth

import (
	"net/http"
	"time"
)

// Auth settings consts
const (
	AuthCookieName    = "AUTH"
	AuthTokenValidity = 72 * time.Hour
)

// Gets authentication cookie with a given token.
//
// Parameters:
//   - value: value to set in the cookie
//   - persistent: if the cookie needs to persist
//
// Returns:
//   - *http.Cookie
func GetAuthCookie(value string, persistent bool) *http.Cookie {

	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	if persistent {
		cookie.Expires = time.Now().Add(AuthTokenValidity)
	}

	return cookie
}

// Gets cookie that invalidates a set authentication cookie.
//
// Returns:
//   - *http.Cookie
func GetAuthDeleteCookie() *http.Cookie {

	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	}

	return cookie
}
