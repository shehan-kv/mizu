package project

type TaskAssigneeSetRequest struct {
	Assignees []int64 `json:"assignees"`
}

func (r *TaskAssigneeSetRequest) Validate() bool {

	return len(r.Assignees) != 0
}
