package repository
import (
	"my-blog/internal/model"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	GetUnreadCount(userId int) (int64, error)
	// [NEW] 新增 Create 方法
	Create(notify *model.Notification) error
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

// [NEW] 实现 Create 方法
func (r *notificationRepository) Create(notify *model.Notification) error {
	return r.db.Create(notify).Error
}