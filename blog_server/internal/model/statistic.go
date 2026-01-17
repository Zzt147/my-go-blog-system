package model

// Statistic 对应 t_statistic 表
type Statistic struct {
	Id             int `gorm:"primaryKey;autoIncrement"`
	ArticleId      int `gorm:"column:article_id"`
	Hits           int `gorm:"column:hits"`         // 点击量/阅读量
	CommentsNum    int `gorm:"column:comments_num"` // 评论数
	Likes          int `gorm:"column:likes"`        // 点赞数
	PrevReadRank   int `gorm:"column:prev_read_rank"`
	ReadRankChange int `gorm:"column:read_rank_change"`
	PrevLikeRank   int `gorm:"column:prev_like_rank"`
	LikeRankChange int `gorm:"column:like_rank_change"`
}

func (Statistic) TableName() string {
	return "t_statistic"
}