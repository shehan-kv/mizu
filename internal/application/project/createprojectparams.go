package project

type CreateProjectParams struct {
	ActorID string
	Name    string
	Status  string
	Members []string
}
