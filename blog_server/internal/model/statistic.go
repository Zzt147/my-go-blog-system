package model

// Statistic 对应 t_statistic 表
type Statistic struct {
	Id          int `gorm:"primaryKey;autoIncrement" json:"id"`
	ArticleId   int `gorm:"column:article_id" json:"articleId"`
	Hits        int `gorm:"column:hits" json:"hits"`         // 点击/阅读量
	Likes       int `gorm:"column:likes" json:"likes"`       // 点赞数
	CommentsNum int `gorm:"column:comments_num" json:"commentsNum"` // 评论数
}

func (Statistic) TableName() string {
	return "t_statistic"
}