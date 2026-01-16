package service
import "my-blog/internal/repository"

type NotificationService interface {
	GetUnreadCount(userId int) (int64, error)
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) GetUnreadCount(userId int) (int64, error) {
	return s.repo.GetUnreadCount(userId)
}