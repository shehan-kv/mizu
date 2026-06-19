package mailer

type ContractEmail struct {
	Subject        string
	RecipientEmail string

	ContractID    string
	ContractName  string
	ContractTerms string
	ProjectID     string
	ProjectName   string

	Signatories []ContractSignatory
}
