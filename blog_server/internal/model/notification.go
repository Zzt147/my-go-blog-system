package model
import "time"

type Notification struct {
	Id        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId    int       `gorm:"column:user_id" json:"userId"` // 接收者
	Content   string    `gorm:"column:content" json:"content"`
	Type      string    `gorm:"column:type" json:"type"`     // 比如 "COMMENT", "LIKE"
	Status    int       `gorm:"column:status" json:"status"` // 0未读 1已读
	Created   time.Time `gorm:"column:created" json:"createDate"`
}

func (Notification) TableName() string {
	return "t_notification"
}