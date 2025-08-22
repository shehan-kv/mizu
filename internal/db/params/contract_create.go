package params

type ContractCreate struct {
	// Name is the human-readable identifier for the contract.
	Name string

	// Version designates the contract’s version string.
	Version string

	// Contract contains the contract definition text.
	Contract string
}
