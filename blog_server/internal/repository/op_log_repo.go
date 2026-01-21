package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

type OpLogRepository interface {
	FindAll(userId int, page, pageSize int) ([]model.OpLog, int64, error)
	Create(opLog *model.OpLog) error
}

type opLogRepository struct {
	db *gorm.DB
}

func NewOpLogRepository(db *gorm.DB) OpLogRepository {
	return &opLogRepository{db: db}
}

// [MODIFY] 修复排序字段
func (r *opLogRepository) FindAll(userId int, page, pageSize int) ([]model.OpLog, int64, error) {
	var logs []model.OpLog
	var total int64
	// 防御性处理 page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 修正：Order("created desc")
	query := r.db.Model(&model.OpLog{}).Where("user_id = ?", userId).Order("created desc")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(pageSize).Offset(offset).Find(&logs).Error
	return logs, total, err
}

func (r *opLogRepository) Create(opLog *model.OpLog) error {
	return r.db.Create(opLog).Error
}
