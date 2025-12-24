package contract

import "time"

type RevisionUser struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type RevisionProject struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type RevisionResponse struct {
	Id           int64           `json:"id"`
	ContractId   int64           `json:"contractId"`
	ContractName string          `json:"contractName"`
	Project      RevisionProject `json:"project"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    *time.Time      `json:"updatedAt"`
	Status       string          `json:"status"`
	ReqUser      RevisionUser    `json:"reqUser"`
	ResUser      *RevisionUser   `json:"resUser"`
}
