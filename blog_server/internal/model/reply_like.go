package model

import "time"

// ReplyLike 对应 t_reply_like 表
type ReplyLike struct {
	Id      int       `gorm:"primaryKey;autoIncrement"`
	UserId  int       `gorm:"column:user_id"`
	ReplyId int       `gorm:"column:reply_id"`
	Created time.Time `gorm:"column:created"`
}

func (ReplyLike) TableName() string {
	return "t_reply_like"
}