package service

import (
	"my-blog/internal/model"
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
	"time"
)

type CommentService interface {
	GetComments(articleId int, pageParams *utils.PageParams) (*utils.Result, error)
	AddComment(comment *model.Comment) error
	GetReplies(commentId int) ([]model.Reply, error)
	AddReply(reply *model.Reply) error
}

type commentService struct {
	repo     repository.CommentRepository
	userRepo repository.UserRepository
}

func NewCommentService(repo repository.CommentRepository, userRepo repository.UserRepository) CommentService {
	return &commentService{repo: repo, userRepo: userRepo}
}

// 获取评论列表
func (s *commentService) GetComments(articleId int, p *utils.PageParams) (*utils.Result, error) {
	comments, total, err := s.repo.GetPageByArticleId(articleId, p.Page, p.Rows)
	if err != nil {
		return nil, err
	}

	// 填充 User 信息
	for i := range comments {
		user, _ := s.userRepo.FindById(comments[i].UserId)
		if user != nil {
			user.Password = ""
			// 前端需要 user 里的 avatar
			comments[i].User = user
			// 如果数据库 author 为空，用 user 表的
			if comments[i].Author == "" {
				comments[i].Author = user.Username
			}
		}
	}

	res := utils.Ok()
	res.Put("comments", comments)
	res.Put("total", total)
	return res, nil
}

// 发表评论 (复刻 Java 逻辑：根据 Author 名字查 UserId)
func (s *commentService) AddComment(comment *model.Comment) error {
	comment.Created = time.Now()
	comment.Status = "approved" // 对应 SQL 默认值
	comment.Likes = 0

	// 关键：如果前端传了 Author 但没传 UserId，我们去查 User 表
	if comment.UserId == 0 && comment.Author != "" {
		user, err := s.userRepo.FindByUsername(comment.Author)
		if err == nil && user != nil {
			comment.UserId = user.Id
		}
	}

	// 反之，如果传了 UserId 但没传 Author (防御性编程)
	if comment.UserId != 0 && comment.Author == "" {
		user, err := s.userRepo.FindById(comment.UserId)
		if err == nil && user != nil {
			comment.Author = user.Username
		}
	}

	return s.repo.Create(comment)
}

// 获取回复
func (s *commentService) GetReplies(commentId int) ([]model.Reply, error) {
	replies, err := s.repo.GetRepliesByCommentId(commentId)
	if err != nil {
		return nil, err
	}

	for i := range replies {
		// 1. 补全回复者信息
		u1, _ := s.userRepo.FindById(replies[i].UserId)
		if u1 != nil {
			u1.Password = ""
			replies[i].User = u1
			
			// [FIX] 确保数据库里的 Author 字段有值
			if replies[i].Author == "" {
				replies[i].Author = u1.Username
			}
			// [FIX] 前端也可能用 username 字段，为了双重保险
			replies[i].Username = u1.Username
		}
		
		// 2. 补全被回复者信息 (TargetName)
		if replies[i].ToUid != 0 {
			u2, _ := s.userRepo.FindById(replies[i].ToUid)
			if u2 != nil {
				u2.Password = ""
				replies[i].TargetUser = u2
				// [FIX] 填充 targetName
				replies[i].TargetAuthor = u2.Username
			}
		} else {
             // 如果是回复层主，TargetAuthor 可能是空的，这没关系
        }
	}
	return replies, nil
}

// 发表回复 (复刻 Java 逻辑：根据 UserId 查 Author 名字)
func (s *commentService) AddReply(reply *model.Reply) error {
	reply.Created = time.Now()
	reply.Likes = 0

	// 1. 根据 UserId 查 Author 名字 (因为前端传的是 ID)
	if reply.UserId != 0 {
		user, err := s.userRepo.FindById(reply.UserId)
		if err == nil && user != nil {
			reply.Author = user.Username
		}
	}

	// 2. 根据 ToUid 查 TargetAuthor 名字 (回复给谁)
	if reply.ToUid != 0 {
		targetUser, err := s.userRepo.FindById(reply.ToUid)
		if err == nil && targetUser != nil {
			reply.TargetAuthor = targetUser.Username
		}
	}

	return s.repo.CreateReply(reply)
}