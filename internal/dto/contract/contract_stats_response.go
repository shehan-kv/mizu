package contract

import "time"

type ContractStatsResponse struct {
	Id                int64         `json:"id"`
	Name              string        `json:"name"`
	ProjectName       *string       `json:"projectName"`
	Status            string        `json:"status"`
	CreatedAt         time.Time     `json:"createdAt"`
	Versions          int64         `json:"versions"`
	Revisions         int64         `json:"numOfRevisions"`
	AcceptedRevisions int64         `json:"acceptedRevisions"`
	LatestVersion     LatestVersion `json:"latestVersion"`
	UserSignature     *string       `json:"userSignature"`
}
