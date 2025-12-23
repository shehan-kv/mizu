package contract

import "time"

type RevisionUser struct {
	FirstName string
	LastName  string
}

type RevisionResponse struct {
	Id           int64         `json:"id"`
	ContractId   int64         `json:"contractId"`
	ContractName string        `json:"contractName"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    *time.Time    `json:"updatedAt"`
	Status       string        `json:"status"`
	ReqUser      RevisionUser  `json:"reqUser"`
	ResUser      *RevisionUser `json:"resUser"`
}
