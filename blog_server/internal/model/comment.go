package model

import "time"

// Comment 对应 t_comment 表
type Comment struct {
	Id        int `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleId int `gorm:"column:article_id" json:"articleId"`

	// [修复] 必须要有 user_id，对应数据库 user_id
	UserId int `gorm:"column:user_id" json:"userId"`

	Content string `gorm:"column:content" json:"content"`

	// [修复] 时间字段跟 Java 保持一致
	Created time.Time `gorm:"column:created" json:"created"`

	Likes int `gorm:"column:likes;default:0" json:"likes"`
	// [修复] 数据库定义 status 是 varchar，但为了兼容之前的逻辑，这里先假设它是 string
	// (注意：你的SQL里 status varchar(200) default 'approved')
	Status string `gorm:"column:status;default:approved" json:"status"`

	// [新增] 对应数据库 author, ip, location
	Author   string `gorm:"column:author" json:"author"`
	Ip       string `gorm:"column:ip" json:"ip"`
	Location string `gorm:"column:location" json:"location"`

	// --- 虚拟字段 ---
	User      *User    `gorm:"-" json:"user"`
	ReplyList []*Reply `gorm:"-" json:"replyList"`

	// --- [NEW] 新增字段 ---
	// gorm:"-" 表示这个字段不映射到数据库表列，只用于后端临时存储传给前端
	ArticleTitle string `gorm:"-" json:"articleTitle"`
}

func (Comment) TableName() string {
	return "t_comment"
}
