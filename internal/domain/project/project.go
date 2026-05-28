package project

import (
	"mizu/internal/domain/common"
	"mizu/internal/domain/iam"
	"time"
)

type Project struct {
	id      ProjectID
	name    Name
	status  Status
	members map[iam.UserID]struct{}

	events []common.Event

	version   int
	createdAt time.Time
	updatedAt time.Time
}

// NewProject is the constructor for project aggregate
func NewProject(
	id ProjectID,
	name Name,
	status Status,
	createdBy iam.UserID,
	members []iam.UserID,
	now time.Time) (*Project, error) {

	// User creating the project is added
	// as a member by default
	memberSet := make(map[iam.UserID]struct{})
	memberSet[createdBy] = struct{}{}

	// Adding the rest of the members
	// without duplicates
	for i := range members {
		if members[i] == "" {
			return nil, ErrProjectMemberIDCannotBeEmpty
		}
		memberSet[members[i]] = struct{}{}
	}

	project := Project{
		id:        id,
		name:      name,
		status:    status,
		version:   1,
		members:   memberSet,
		createdAt: now,
		updatedAt: now,
	}

	project.events = append(project.events, ProjectCreatedEvent{
		ProjectID:  id,
		Name:       name,
		Members:    project.Members(),
		OccurredAt: now,
	})

	return &project, nil
}

func RestoreProject(
	id ProjectID,
	name Name,
	status Status,
	members []iam.UserID,
	version int,
	created_at time.Time,
	updated_at time.Time,
) *Project {

	mMap := make(map[iam.UserID]struct{}, len(members))
	for i := range members {
		mMap[members[i]] = struct{}{}
	}

	return &Project{
		id:        id,
		name:      name,
		status:    status,
		members:   mMap,
		version:   version,
		createdAt: created_at,
		updatedAt: updated_at,
	}

}

// ID returns the project id
func (p *Project) ID() ProjectID {
	return p.id
}

// Name returns the project name
func (p *Project) Name() Name {
	return p.name
}

// Status returns the project status
func (p *Project) Status() Status {
	return p.status
}

// Members returns a list of current project members
func (p *Project) Members() []iam.UserID {

	// Pre-allocate the slice to avoid the dynamic growth that occurs
	// when using slices.Collect(maps.Keys(p.members)).
	m := make([]iam.UserID, len(p.members))

	i := 0
	for id := range p.members {
		m[i] = id
		i++
	}

	return m
}

func (p *Project) Version() int {
	return p.version
}

func (p *Project) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Project) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Project) AcceptsTasks() bool {
	return p.status != StatusCancelled && p.status != StatusCompleted
}

func (p *Project) MarkAsStarted(now time.Time) {
	p.status = StatusStarted
	p.updatedAt = now
}

func (p *Project) MarkAsPaused(now time.Time) {
	p.status = StatusPaused
	p.updatedAt = now
}

func (p *Project) MarkAsCancelled(now time.Time) {
	p.status = StatusCancelled
	p.updatedAt = now
}

func (p *Project) MarkAsCompleted(now time.Time) {
	p.status = StatusCompleted
	p.updatedAt = now
}

func (p *Project) ReplaceMembers(members []iam.UserID, now time.Time) error {

	if len(members) == 0 {
		return ErrProjectMustHaveAtLeastOneMember
	}

	m := make(map[iam.UserID]struct{}, len(members))
	for i := range members {
		if members[i] == "" {
			return ErrProjectInvalidMember
		}

		m[members[i]] = struct{}{}
	}

	p.members = m
	p.updatedAt = now

	return nil
}

func (p *Project) AddMember(memberID iam.UserID, now time.Time) error {
	if memberID == "" {
		return ErrProjectMemberIDCannotBeEmpty
	}

	p.members[memberID] = struct{}{}
	p.updatedAt = now

	return nil
}

func (p *Project) RemoveMember(memberID iam.UserID, now time.Time) error {
	if memberID == "" {
		return ErrProjectMemberIDCannotBeEmpty
	}

	delete(p.members, memberID)
	p.updatedAt = now

	return nil
}

func (p *Project) HasMember(memberID iam.UserID) bool {
	_, ok := p.members[memberID]
	return ok
}

func (p *Project) PullEvents() []common.Event {
	events := p.events
	p.events = nil

	return events
}

func (p *Project) Equals(prj Project) bool {
	return p.id == prj.ID()
}
