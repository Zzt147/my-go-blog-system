package repository
import (
	"my-blog/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	GetUnreadCount(userId int) (int64, error)
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) GetUnreadCount(userId int) (int64, error) {
	var count int64
	// 假设 status=0 是未读
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND status = 0", userId).
		Count(&count).Error
	return count, err
}