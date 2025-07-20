package middleware

import (
	"net/http"
	"slices"
)

// A chain of middleware functions
type Chain struct {
	mws []func(http.HandlerFunc) http.HandlerFunc
}

// Creates a new instance of the middleware chain
//
// Returns:
//   - a pointer to a new Chain
func NewChain() *Chain {
	return &Chain{mws: nil}
}

// Adds a list of middleware functions to the chain
//
// Parameters:
//   - mws: a comma separated list of middleware (ex: mwOne, mwTwo, mwThree)
func (c *Chain) Add(mws ...func(http.HandlerFunc) http.HandlerFunc) {
	c.mws = append(c.mws, mws...)
}

// Wraps a handler function in the middleware chain
//
// Parameters:
//   - handler: a handler function that implements http.HandlerFunc
//
// Returns:
//   - an http.HandlerFunc wrapped in the middleware chain
func (c *Chain) Handle(handler http.HandlerFunc) http.HandlerFunc {

	for _, mw := range slices.Backward(c.mws) {
		handler = mw(handler)
	}

	return handler
}
