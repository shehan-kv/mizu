package project

type ProjectID string

func NewProjectID(id string) (ProjectID, error) {
	if id == "" {
		return "", ErrProjectIDCannotBeEmpty
	}

	return ProjectID(id), nil
}

func (p ProjectID) String() string {
	return string(p)
}
