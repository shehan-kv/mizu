package file

import "time"

type FileResponse struct {
	Id           int64            `json:"id"`
	ChannelId    int64            `json:"channelId"`
	User         FileUserResponse `json:"user"`
	OriginalName string           `json:"originalName"`
	SavedName    string           `json:"savedName"`
	UploadedAt   time.Time        `json:"uploadedAt"`
	Url          string           `json:"url"`
	Size         int64            `json:"size"`
}
