package model

import "time"

// Reply 对应 t_reply 表
type Reply struct {
	Id           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentId    int       `gorm:"column:comment_id" json:"commentId"`
	UserId       int       `gorm:"column:user_id" json:"userId"`
	
	// [新增] 对应前端传来的 toUid
	ToUid        int       `gorm:"column:to_uid" json:"toUid"` 
	
	Content      string    `gorm:"column:content" json:"content"`
	Created      time.Time `gorm:"column:created" json:"created"`
	
	// [新增] 数据库里的字段
	Author       string    `gorm:"column:author" json:"author"`
	TargetAuthor string    `gorm:"column:target_author" json:"targetName"` // 前端叫 targetName
	Ip           string    `gorm:"column:ip" json:"ip"`
	Location     string    `gorm:"column:location" json:"location"`
	Likes        int       `gorm:"column:likes;default:0" json:"likes"`

	// --- 虚拟字段 ---
	User       *User `gorm:"-" json:"user"`
	TargetUser *User `gorm:"-" json:"targetUser"`

	// 兼容前端直接读取 username (虽然 User 对象里也有，但前端可能有 legacy 代码)
	Username   string `gorm:"-" json:"username"`
}

func (Reply) TableName() string {
	return "t_reply"
}