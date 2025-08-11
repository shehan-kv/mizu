package params

// Represents a user verify email request
// for sending an email with the verify token
type VerifyRequest struct {
	FirstName     string
	LastName      string
	Email         string
	Token         string
	CorrelationId string
}
