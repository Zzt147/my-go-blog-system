package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll() ([]*model.Category, error)
	FindById(id int) (*model.Category, error)
	FindByParentId(parentId int) ([]*model.Category, error)
	Create(category *model.Category) error
	Update(category *model.Category) error
	Delete(id int) error
	// [NEW] 批量更新 (用于拖拽排序)
	UpdateBatch(categories []model.Category) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll() ([]*model.Category, error) {
	var list []*model.Category
	// 按 sort 排序
	err := r.db.Order("sort asc, id asc").Find(&list).Error
	return list, err
}

func (r *categoryRepository) FindById(id int) (*model.Category, error) {
	var category model.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindByParentId(parentId int) ([]*model.Category, error) {
	var list []*model.Category
	err := r.db.Where("parent_id = ?", parentId).Order("sort asc").Find(&list).Error
	return list, err
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) Update(category *model.Category) error {
	// 使用 Updates 更新指定字段，或者 Select("*") Save
	// 这里为了简单和安全，只更新 Name, Sort, ParentId
	return r.db.Model(category).Updates(map[string]interface{}{
		"name":      category.Name,
		"sort":      category.Sort,
		"parent_id": category.ParentId,
	}).Error
}

func (r *categoryRepository) Delete(id int) error {
	return r.db.Delete(&model.Category{}, id).Error
}

// [NEW] 批量更新 (事务处理)
func (r *categoryRepository) UpdateBatch(categories []model.Category) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, cat := range categories {
			if err := tx.Model(&model.Category{Id: cat.Id}).
				Updates(map[string]interface{}{
					"parent_id": cat.ParentId,
					"sort":      cat.Sort,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
