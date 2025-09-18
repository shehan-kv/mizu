package aggregates

import "time"

type FileWithUser struct {
	Id            int64
	ChannelId     int64
	UserId        int64
	UserFirstName string
	UserLastName  string
	OriginalName  string
	SavedName     string
	UploadedAt    time.Time
	Url           string
	Size          int64
}
