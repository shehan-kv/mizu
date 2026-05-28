package message

import "errors"

var (
	ErrChannelMustHaveAtLeastTwoMembers = errors.New("channel must have at least two members")
	ErrChannelConcurrentModification    = errors.New("channel concurrent modification")
	ErrChannelMembersMustBeUnique       = errors.New("channel members must be unique")
	ErrChannelMemberAlreadyExists       = errors.New("channel member already exists")
	ErrChannelMustHaveStaffMember       = errors.New("channel must have staff member")
	ErrChannelNameCannotBeEmpty         = errors.New("channel name cannot be empty")
	ErrChannelIDCannotBeEmpty           = errors.New("channel id cannot be empty")
	ErrChannelMemberNotFound            = errors.New("channel member not found")
	ErrNotChannelMember                 = errors.New("channel not channel member")
	ErrChannelNotFound                  = errors.New("channel not found")

	ErrMessageContentCannotBeEmpty = errors.New("message content cannot be empty")
	ErrMessageIDCannotBeEmpty      = errors.New("message id cannot be empty")
	ErrMessageContentTooLong       = errors.New("message content too long")
	ErrMessageInvalidSender        = errors.New("message invalid sender")
	ErrMessageNotFound             = errors.New("message not found")

	ErrFileIDCannotBeEmpty = errors.New("file id cannot be empty")
)
