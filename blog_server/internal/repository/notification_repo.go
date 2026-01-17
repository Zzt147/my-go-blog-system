package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(notify *model.Notification) error
	// [NEW] 分页获取我的通知
	GetPage(receiverId, page, rows int) ([]*model.Notification, int64, error)
	// [NEW] 获取未读数量
	GetUnreadCount(receiverId int) (int64, error)
	// [NEW] 标记单条为已读
	UpdateStatus(id int, status int) error
	// [NEW] 标记所有为已读
	MarkAllRead(receiverId int) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notify *model.Notification) error {
	return r.db.Create(notify).Error
}

// 分页获取
func (r *notificationRepository) GetPage(receiverId, page, rows int) ([]*model.Notification, int64, error) {
	var list []*model.Notification
	var total int64
	offset := (page - 1) * rows

	// 1. 查总数
	query := r.db.Model(&model.Notification{}).Where("receiver_id = ?", receiverId)
	query.Count(&total)

	// 2. 查列表 (未读优先，然后按时间倒序)
	// status=0 是未读，status=1 是已读。我们希望未读在上面 -> Asc排序即可(0在1前)
	// 或者直接按 Created desc
	err := query.
		Order("status asc"). // 未读在前
		Order("created desc"). // 新的在前
		Limit(rows).Offset(offset).
		Find(&list).Error

	return list, total, err
}

// 获取未读数
func (r *notificationRepository) GetUnreadCount(receiverId int) (int64, error) {
	var count int64
	// status = 0 代表未读
	err := r.db.Model(&model.Notification{}).
		Where("receiver_id = ? AND status = ?", receiverId, 0).
		Count(&count).Error
	return count, err
}

// 标记单条已读
func (r *notificationRepository) UpdateStatus(id int, status int) error {
	return r.db.Model(&model.Notification{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// 标记全部已读
func (r *notificationRepository) MarkAllRead(receiverId int) error {
	return r.db.Model(&model.Notification{}).
		Where("receiver_id = ? AND status = ?", receiverId, 0).
		Update("status", 1).Error
}
