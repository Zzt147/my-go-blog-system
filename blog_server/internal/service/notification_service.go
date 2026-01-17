package service

import (
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
)

type NotificationService interface {
	GetPage(userId, page, rows int) (*utils.Result, error)
	GetUnreadCount(userId int) (int64, error)
	MarkAsRead(id int) error
	MarkAllAsRead(userId int) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

// 获取分页
func (s *notificationService) GetPage(userId, page, rows int) (*utils.Result, error) {
	list, total, err := s.repo.GetPage(userId, page, rows)
	if err != nil {
		return nil, err
	}
	res := utils.Ok()
	res.Put("data", list) // 前端可能是取 res.data.map.data
	res.Put("total", total)
	return res, nil
}

// 获取未读数
func (s *notificationService) GetUnreadCount(userId int) (int64, error) {
	return s.repo.GetUnreadCount(userId)
}

// 标记单条
func (s *notificationService) MarkAsRead(id int) error {
	// status: 1 = 已读
	return s.repo.UpdateStatus(id, 1)
}

// 标记所有
func (s *notificationService) MarkAllAsRead(userId int) error {
	return s.repo.MarkAllRead(userId)
}
