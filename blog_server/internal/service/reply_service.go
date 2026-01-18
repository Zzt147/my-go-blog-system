package service

import (
	"errors"
	"fmt"
	"my-blog/internal/model"
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
	"time"
)

type ReplyService interface {
	GetReplies(commentId int) ([]model.Reply, error)
	AddReply(reply *model.Reply) error
	LikeReply(userId, replyId int) (string, error)
}

type replyService struct {
	repo        repository.ReplyRepository
	userRepo    repository.UserRepository
	commentRepo repository.CommentRepository
	notifyRepo  repository.NotificationRepository
	articleRepo repository.ArticleRepository
}

func NewReplyService(
	repo repository.ReplyRepository,
	userRepo repository.UserRepository,
	commentRepo repository.CommentRepository,
	notifyRepo repository.NotificationRepository,
	articleRepo repository.ArticleRepository,
) ReplyService {
	return &replyService{
		repo:        repo,
		userRepo:    userRepo,
		commentRepo: commentRepo,
		notifyRepo:  notifyRepo,
		articleRepo: articleRepo,
	}
}

func (s *replyService) GetReplies(commentId int) ([]model.Reply, error) {
	replies, err := s.repo.GetRepliesByCommentId(commentId)
	if err != nil {
		return nil, err
	}

	for i := range replies {
		u1, _ := s.userRepo.FindById(replies[i].UserId)
		if u1 != nil {
			u1.Password = ""
			replies[i].User = u1
			replies[i].Username = u1.Username
			if replies[i].Author == "" {
				replies[i].Author = u1.Username
			}
		}
		if replies[i].ToUid != 0 {
			u2, _ := s.userRepo.FindById(replies[i].ToUid)
			if u2 != nil {
				u2.Password = ""
				replies[i].TargetUser = u2
				replies[i].TargetAuthor = u2.Username
			}
		}
	}
	return replies, nil
}

func (s *replyService) AddReply(reply *model.Reply) error {
	// [FIX] 使用 FindById
	parentComment, err := s.commentRepo.FindById(reply.CommentId)
	if err != nil || parentComment == nil {
		return errors.New("父评论不存在")
	}

	reply.Created = time.Now()
	reply.Likes = 0

	if reply.UserId != 0 {
		user, _ := s.userRepo.FindById(reply.UserId)
		if user != nil {
			reply.Author = user.Username
		}
	}
	if reply.ToUid != 0 {
		target, _ := s.userRepo.FindById(reply.ToUid)
		if target != nil {
			reply.TargetAuthor = target.Username
		}
	}

	if err := s.repo.CreateReply(reply); err != nil {
		return err
	}

	go func() {
		receiverId := reply.ToUid
		if receiverId == 0 {
			receiverId = parentComment.UserId
		}

		if receiverId != 0 && receiverId != reply.UserId {
			title := ""
			article, _ := s.articleRepo.FindById(parentComment.ArticleId)
			if article != nil {
				title = article.Title
			}

			notify := &model.Notification{
				ReceiverId: receiverId,
				SenderId:   reply.UserId,
				SenderName: reply.Author,
				ArticleId:  parentComment.ArticleId,
				CommentId:  reply.CommentId,
				Content:    fmt.Sprintf("在《%s》回复了你: %s", utils.SubString(title, 10), utils.SubString(reply.Content, 20)),
				Type:       "REPLY",
				Status:     0,
				Created:    time.Now(),
			}
			s.notifyRepo.Create(notify)
		}
	}()
	return nil
}

func (s *replyService) LikeReply(userId, replyId int) (string, error) {
	like, _ := s.repo.FindReplyLike(userId, replyId)
	if like != nil && like.Id > 0 {
		s.repo.DeleteReplyLike(userId, replyId)
		s.repo.UpdateReplyLikesCount(replyId, -1)
		return "取消点赞", nil
	}
	s.repo.AddReplyLike(&model.ReplyLike{UserId: userId, ReplyId: replyId, Created: time.Now()})
	s.repo.UpdateReplyLikesCount(replyId, 1)
	return "点赞成功", nil
}
