package params

type FileCreate struct {
	ChannelId    int64
	UserId       int64
	OriginalName string
	SavedName    string
	Url          string
	Size         int64
}
