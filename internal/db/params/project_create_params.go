package params

// Parameters to create a project
type ProjectCreateParams struct {
	Name    string
	Status  string
	Members []int64
}
