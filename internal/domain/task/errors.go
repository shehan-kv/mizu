package task

import "errors"

var (
	ErrTaskAssigneeNotProjectMember = errors.New("task assignee is not a project member")
	ErrTaskConcurrentModification   = errors.New("task concurrent modification")
	ErrTaskMinutesMustBePositive    = errors.New("task minutes must be positive")
	ErrTaskNameCannotBeEmpty        = errors.New("task name cannot be empty")
	ErrTaskAlreadyInProgress        = errors.New("task is already in progress")
	ErrTaskNotAssignedToUser        = errors.New("task is not assigned to user")
	ErrTaskAssigneeNotFound         = errors.New("task assignee not found")
	ErrTaskAlreadyCompleted         = errors.New("task is already completed")
	ErrTaskIDCannotBeEmpty          = errors.New("task id cannot be empty")
	ErrTaskInvalidPriority          = errors.New("task invalid priority")
	ErrTaskAlreadyBacklog           = errors.New("task is already in backlog")
	ErrTaskInvalidStatus            = errors.New("task invalid status")
	ErrTaskNotFound                 = errors.New("task not found")
)
