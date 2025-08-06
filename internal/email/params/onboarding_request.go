package params

// Represents an onboarding email request
// for sending an email with the onboarding token
type OnboardingRequest struct {
	FirstName     string
	LastName      string
	Email         string
	Token         string
	CorrelationId string
}
