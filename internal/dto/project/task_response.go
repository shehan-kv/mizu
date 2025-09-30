package project

import "time"

type TaskResponse struct {
	Id             int64                  `json:"id"`
	ProjectId      int64                  `json:"projectId"`
	Name           string                 `json:"name"`
	Status         string                 `json:"status"`
	Priority       string                 `json:"priority"`
	Description    string                 `json:"description"`
	CreatedAt      time.Time              `json:"createdAt"`
	EstTimeMinutes int64                  `json:"estTimeMinutes"`
	Assignees      []TaskAssigneeResponse `json:"assignees"`
}
