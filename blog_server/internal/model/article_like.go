package model

import "time"

// ArticleLike 对应 t_article_like 表
type ArticleLike struct {
	Id        int       `gorm:"primaryKey;autoIncrement"`
	UserId    int       `gorm:"column:user_id"`
	ArticleId int       `gorm:"column:article_id"`
	Created   time.Time `gorm:"column:created"`
}

func (ArticleLike) TableName() string {
	return "t_article_like"
}