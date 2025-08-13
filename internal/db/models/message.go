package models

import "time"

type Message struct {
	Id        int64
	ChannelId int64
	UserId    int64
	CreatedAt time.Time
	Type      string
}
