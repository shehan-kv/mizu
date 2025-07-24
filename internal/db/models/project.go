package models

import "database/sql"

// Represents a project in database
type Project struct {
	Id        int64
	Name      string
	Status    int64
	CreatedAt sql.NullTime
}
