package model

import "time"

type Notification struct {
	Id int `gorm:"primaryKey;autoIncrement" json:"id"`

	// ✅ [修正] 对应 SQL 中的 receiver_id (接收者)
	// 注意：JSON 依然可以叫 userId 或者 receiverId，看你个人习惯，这里为了兼容旧代码暂用 userId
	ReceiverId int `gorm:"column:receiver_id" json:"receiverId"`

	// ✅ [新增] 对应 SQL 中的 sender_id (发送者)
	SenderId int `gorm:"column:sender_id" json:"senderId"`

	// ✅ [新增] 对应 SQL 中的 sender_name (发送者名字，用于前端显示)
	SenderName string `gorm:"column:sender_name" json:"senderName"`

	// ✅ [新增] 对应 SQL 中的 article_id (用于跳转)
	ArticleId int `gorm:"column:article_id" json:"articleId"`

	// ✅ [新增] 对应 SQL 中的 comment_id (用于跳转定位)
	CommentId int `gorm:"column:comment_id" json:"commentId"`

	Content string    `gorm:"column:content" json:"content"`
	Type    string    `gorm:"column:type" json:"type"`     // COMMENT, REPLY, LIKE
	Status  int       `gorm:"column:status" json:"status"` // 0:未读 1:已读
	Created time.Time `gorm:"column:created" json:"created"`
}

func (Notification) TableName() string {
	return "t_notification"
}
