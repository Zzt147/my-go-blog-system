package model

import "time"

// CommentLike 对应 t_comment_like 表
type CommentLike struct {
	Id        int       `gorm:"primaryKey;autoIncrement"`
	UserId    int       `gorm:"column:user_id"`
	CommentId int       `gorm:"column:comment_id"`
	Created   time.Time `gorm:"column:created"`
}

func (CommentLike) TableName() string {
	return "t_comment_like"
}