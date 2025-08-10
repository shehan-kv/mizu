package params

// Parameters to create a project
type ProjectCreate struct {
	Name    string
	Status  string
	Members []int64
}
