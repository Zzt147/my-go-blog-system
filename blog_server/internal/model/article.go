package model

import "time"

// Article 对应数据库 t_article 表
type Article struct {
	Id           int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Title        string     `gorm:"column:title" json:"title"`
	Author       string     `gorm:"column:author" json:"author"`
	Content      string     `gorm:"column:content" json:"content"`
	Created      time.Time  `gorm:"column:created" json:"created"`
	Modified     *time.Time `gorm:"column:modified" json:"modified"`
	Categories   string     `gorm:"column:categories" json:"categories"` // 对应 varchar
	Tags         string     `gorm:"column:tags" json:"tags"`
	AllowComment int        `gorm:"column:allow_comment" json:"allowComment"` // tinyint(1) 建议用 int 或 bool
	Thumbnail    string     `gorm:"column:thumbnail" json:"thumbnail"`
	UserId       int        `gorm:"column:user_id" json:"userId"`
	Location     string     `gorm:"column:location" json:"location"`

	// --- 👇 下面是仅仅为了返回给前端用的“虚字段” (对应 Java 的 @TableField(exist=false)) ---
	// gorm:"-" 表示 GORM 读写数据库时忽略它
	// gorm:"->" 表示只读 (Scan 时可以写入，但 Save 时不保存)，这里我们用 "-" 手动填充更稳妥
	Likes      int    `gorm:"->" json:"likes"`     // 点赞数
	Views      int    `gorm:"->" json:"views"`     // 虚拟字段 (对应数据库的 hits)
	AuthorName string `gorm:"-" json:"authorName"` // 作者昵称

	// isLiked 是纯业务字段，数据库完全没有，还是保持 gorm:"-"
	IsLiked bool `gorm:"-" json:"isLiked"`
}

// [NEW] 文章查询条件 (对应 Java 的 ArticleCondition)
type ArticleCondition struct {
	Tag        string `json:"tag"`
	CategoryId int    `json:"categoryId"`
	Title      string `json:"title"`
	Content    string `json:"content"`

	// [NEW] 新增用户ID筛选 (用于"我的文章")
	UserId int `json:"userId"`
}

// TableName 指定表名为 t_article
func (Article) TableName() string {
	return "t_article"
}
