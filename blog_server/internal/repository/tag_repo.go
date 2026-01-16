package repository

import (
	"my-blog/internal/model"
	"gorm.io/gorm"
)

type TagRepository interface {
	GetAllTags() ([]model.Tag, error)
	// 复刻 Java: SELECT t.id, t.name, COUNT(at.tag_id) as count ...
  GetHotTags(limit int) ([]model.Tag, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) GetAllTags() ([]model.Tag, error) {
	var tags []model.Tag
	// 对应 SQL: SELECT * FROM t_tag
	result := r.db.Find(&tags)
	return tags, result.Error
}

func (r *tagRepository) GetHotTags(limit int) ([]model.Tag, error) {
	var tags []model.Tag
	
	// 直接使用原生 SQL，这跟你的 Java MyBatis 注解一模一样
	sql := `
		SELECT t.id, t.name, COUNT(at.tag_id) as count 
		FROM t_tag t 
		LEFT JOIN t_article_tag at ON t.id = at.tag_id 
		GROUP BY t.id 
		ORDER BY count DESC 
		LIMIT ?
	`
	// Raw 执行原生 SQL，Scan 映射结果
	err := r.db.Raw(sql, limit).Scan(&tags).Error
	return tags, err
}