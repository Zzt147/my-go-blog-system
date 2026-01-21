package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

type TagRepository interface {
	GetAllTags() ([]model.Tag, error)
	// 复刻 Java: SELECT t.id, t.name, COUNT(at.tag_id) as count ...
	GetHotTags(limit int) ([]model.Tag, error)

	// [NEW] 分页查询
	GetPage(page, pageSize int) ([]model.Tag, int64, error)
	// [NEW] 更新
	Update(tag *model.Tag) error
	// [NEW] 删除
	Delete(id int) error
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

// [NEW] 分页查询实现
func (r *tagRepository) GetPage(page, pageSize int) ([]model.Tag, int64, error) {
	var tags []model.Tag
	var total int64
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := r.db.Model(&model.Tag{}).Order("id desc")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(pageSize).Offset(offset).Find(&tags).Error
	return tags, total, err
}

// [NEW] 更新
func (r *tagRepository) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// [NEW] 删除
func (r *tagRepository) Delete(id int) error {
	return r.db.Delete(&model.Tag{}, id).Error
}
