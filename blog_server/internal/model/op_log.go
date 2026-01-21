package model

import "time"

// [MODIFY] 严格对应数据库 t_op_log 表结构
type OpLog struct {
	Id       int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId   int       `gorm:"column:user_id" json:"userId"`
	Type     string    `gorm:"column:type" json:"type"`       // BROWSE, COMMENT, LOGIN
	Content  string    `gorm:"column:content" json:"content"` // 操作描述
	TargetId int       `gorm:"column:target_id" json:"targetId"`
	Created  time.Time `gorm:"column:created" json:"created"` // 操作时间
}

func (OpLog) TableName() string {
	return "t_op_log"
}
