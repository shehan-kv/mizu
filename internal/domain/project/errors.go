package project

import "errors"

var (
	ErrProjectCannotRemoveUnassignedMember = errors.New("project cannot remove unassigned member")
	ErrProjectMustHaveAtLeastOneMember     = errors.New("project must have at least one member")
	ErrProjectCannotRemoveLastMember       = errors.New("project cannot remove last member")
	ErrProjectConcurrentModification       = errors.New("project concurrent modification")
	ErrProjectMustHaveAdministrator        = errors.New("project must have an administrator")
	ErrProjectMemberIDCannotBeEmpty        = errors.New("project member id cannot be empty")
	ErrProjectNameCannotBeEmpty            = errors.New("project name cannot be empty")
	ErrProjectCreatorIsRequired            = errors.New("project creator is required")
	ErrProjectNotAcceptingTasks            = errors.New("project is not accepting tasks")
	ErrProjectNameAlreadyExists            = errors.New("project name already exists")
	ErrProjectIDCannotBeEmpty              = errors.New("project id cannot be empty")
	ErrNotProjectMember                    = errors.New("member is not in project")
	ErrProjectInvalidStatus                = errors.New("project invalid status")
	ErrProjectInvalidMember                = errors.New("project invalid member")
	ErrProjectNotFound                     = errors.New("project not found")
)
