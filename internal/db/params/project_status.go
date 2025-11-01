package params

type ProjectStatus = string

const (
	ProjectStatusStarted   ProjectStatus = "started"
	ProjectStatusPaused    ProjectStatus = "paused"
	ProjectStatusCancelled ProjectStatus = "cancelled"
	ProjectStatusCompleted ProjectStatus = "completed"
)
