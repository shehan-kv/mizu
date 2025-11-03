package project

type MemberSetRequest struct {
	Members []int64 `json:"members"`
}

func (r *MemberSetRequest) Validate() bool {

	return len(r.Members) != 0
}
