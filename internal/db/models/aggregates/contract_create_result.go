package aggregates

type ContractCreateResult struct {

	// ContractId is the ID of the newly generated contract
	ContractId int64

	// Version designates the contract’s version string.
	Version string

	// Messages map channel IDs to system generated messages for creating
	// a new contract. This could be utilized by the caller to stream messages
	// to relevant clients.
	Messages map[int64]MessageWithUser

	// UserIds map channel IDs to user IDs in the channel. Useful when sending
	// events to all connected users assigned to the specific channel.
	UserIds map[int64][]int64
}
