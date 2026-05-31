package cookie

import (
	"net/http"
	"time"
)

// AuthCookie creates authentication cookies with preset configuration.
type AuthCookie struct {

	// Auth cookie settings
	name     string
	validity time.Duration
}

// NewAuthCookie returns an AuthCookie with sensible defaults.
func NewAuthCookie() AuthCookie {
	return AuthCookie{
		name:     "AUTH",
		validity: 72 * time.Hour,
	}
}

// New returns an authentication cookie with the given token as its value.
// If persistent is true, the cookie will expire after the configured validity duration.
// Otherwise the cookie will be session-scoped and expire when the browser closes.
func (c AuthCookie) New(token string, persistent bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:     c.name,
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}
	if persistent {
		cookie.Expires = time.Now().Add(c.validity)
	}
	return cookie
}

// Extract returns the authentication token associated with the request.
// If no token is available, nil is returned.
func (c AuthCookie) Extract(r *http.Request) *string {
	cookie, err := r.Cookie(c.name)
	if err != nil {
		return nil
	}

	if cookie.Value == "" {
		return nil
	}

	return &cookie.Value
}

func (c AuthCookie) Validity() time.Duration {
	return c.validity
}

// Expire returns an expired cookie that instructs the browser to delete
// the authentication cookie.
func (c AuthCookie) Expire() *http.Cookie {
	return &http.Cookie{
		Name:     c.name,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	}
}

func (c AuthCookie) CookieName() string {
	return c.name
}
