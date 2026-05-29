package lifecycle

import "context"

// Starter is implemented by components that require
// a background goroutine before they can process work.
type Starter interface {
	Start(ctx context.Context)
}

// Stopper is implemented by components that must drain
// in-flight work before the process exits.
type Stopper interface {
	Stop()
}
