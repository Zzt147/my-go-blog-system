package service

import (
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
)

type OpLogService interface {
	GetMyFootprints(userId int, pageParams *utils.PageParams) (*utils.Result, error)
}

type opLogService struct {
	repo repository.OpLogRepository
}

func NewOpLogService(repo repository.OpLogRepository) OpLogService {
	return &opLogService{repo: repo}
}

// [NEW] 获取我的足迹
func (s *opLogService) GetMyFootprints(userId int, p *utils.PageParams) (*utils.Result, error) {
	logs, total, err := s.repo.FindAll(userId, p.Page, p.Rows)
	if err != nil {
		return nil, err
	}

	res := utils.Ok()
	res.Put("opLogs", logs) // 对应 Java: map.put("opLogs", ...)
	res.Put("total", total)
	return res, nil
}
