package message

import (
	"mizu/internal/domain/iam"
	"mizu/internal/domain/project"
	"slices"
	"time"
)

type Channel struct {
	id        ChannelID
	projectID *project.ProjectID
	name      ChannelName
	members   []iam.UserID

	version   int
	createdAt time.Time
	updatedAt time.Time
}

func NewChannel(
	id ChannelID,
	projectID *project.ProjectID,
	name ChannelName,
	members []iam.UserID,
	now time.Time,
) (*Channel, error) {

	if len(members) < 2 {
		return nil, ErrChannelMustHaveAtLeastTwoMembers
	}
	if hasDuplicateMembers(members) {
		return nil, ErrChannelMembersMustBeUnique
	}

	return &Channel{
		id:        id,
		projectID: projectID,
		name:      name,
		members:   members,

		version:   1,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreChannel(
	id ChannelID,
	projectID *project.ProjectID,
	name ChannelName,
	members []iam.UserID,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) *Channel {

	return &Channel{
		id:        id,
		projectID: projectID,
		name:      name,
		members:   members,
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func NewProjectChannel(
	id ChannelID,
	projectID project.ProjectID,
	name ChannelName,
	members []iam.UserID,
	now time.Time,
) (*Channel, error) {
	return NewChannel(id, &projectID, name, members, now)
}

func NewUserChannel(
	id ChannelID,
	name ChannelName,
	members []iam.UserID,
	now time.Time,
) (*Channel, error) {
	return NewChannel(id, nil, name, members, now)
}

func hasDuplicateMembers(members []iam.UserID) bool {
	seen := make(map[iam.UserID]struct{}, len(members))
	for i := range members {
		if _, exists := seen[members[i]]; exists {
			return true
		}
		seen[members[i]] = struct{}{}
	}
	return false
}

func (c *Channel) ID() ChannelID {
	return c.id
}

func (c *Channel) ProjectID() *project.ProjectID {
	return c.projectID
}

func (c *Channel) Name() ChannelName {
	return c.name
}

func (c *Channel) Version() int {
	return c.version
}

func (c *Channel) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Channel) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Channel) HasMember(userID iam.UserID) bool {
	return slices.Contains(c.members, userID)
}

func (c *Channel) AddMember(userID iam.UserID, now time.Time) error {
	if c.HasMember(userID) {
		return ErrChannelMemberAlreadyExists
	}

	c.members = append(c.members, userID)
	c.updatedAt = now

	return nil
}

func (c *Channel) RemoveMember(userID iam.UserID, now time.Time) error {
	if len(c.members) <= 2 {
		return ErrChannelMustHaveAtLeastTwoMembers
	}
	for i := range c.members {
		if c.members[i] == userID {
			c.members = append(c.members[:i], c.members[i+1:]...)
			c.updatedAt = now

			return nil
		}
	}
	return ErrChannelMemberNotFound
}

func (c *Channel) ReplaceMembers(userIDs []iam.UserID, now time.Time) error {
	if len(userIDs) < 2 {
		return ErrChannelMustHaveAtLeastTwoMembers
	}
	if hasDuplicateMembers(userIDs) {
		return ErrChannelMembersMustBeUnique
	}

	c.members = userIDs
	c.updatedAt = now

	return nil
}

func (c *Channel) Members() []iam.UserID {
	members := make([]iam.UserID, len(c.members))
	copy(members, c.members)
	return members
}
