package service

import (
	"fmt"
	"my-blog/internal/model"
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
	"time"
)

type CommentService interface {
	GetComments(articleId int, pageParams *utils.PageParams) (*utils.Result, error)
	AddComment(comment *model.Comment) error
	// [NEW] 点赞
	LikeComment(userId, commentId int) (string, error) // 返回 "点赞成功" 或 "取消点赞"
}

type commentService struct {
	repo     repository.CommentRepository
	userRepo repository.UserRepository

	// [NEW] 注入通知服务和文章Repo (为了查文章作者)
	notifyRepo  repository.NotificationRepository
	articleRepo repository.ArticleRepository

	replyRepo repository.ReplyRepository
}

// [MODIFIED] 修改构造函数，注入新的依赖
func NewCommentService(
	repo repository.CommentRepository,
	userRepo repository.UserRepository,
	notifyRepo repository.NotificationRepository, // 新增
	articleRepo repository.ArticleRepository,     // 新增
	replyRepo repository.ReplyRepository,
) CommentService {
	return &commentService{
		repo:        repo,
		userRepo:    userRepo,
		notifyRepo:  notifyRepo,
		articleRepo: articleRepo,
		replyRepo:   replyRepo,
	}
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

	comment.Created = time.Now()
	comment.Status = "approved"

	// 1. 存评论
	if err := s.repo.Create(comment); err != nil {
		return err
	}

	// 2. [NEW] ✅ 统计数 +1 (异步执行即可，不影响主流程)
	go s.repo.UpdateArticleCommentCount(comment.ArticleId, 1)

	// 2. [NEW] 发送通知 (复刻 NotificationAspect)
	go func() { // 开个协程异步发，不卡主线程
		article, _ := s.articleRepo.FindById(comment.ArticleId)
		if article != nil && article.UserId != comment.UserId { // 自己评论自己不发通知
			notify := &model.Notification{
				ReceiverId: article.UserId,    // 接收者：文章作者
				SenderId:   comment.UserId,    // ✅ 必须填
				SenderName: comment.Author,    // ✅ 必须填
				ArticleId:  comment.ArticleId, // ✅ 必须填，否则前端跳不过去
				CommentId:  comment.Id,        // ✅ 必须填
				Content:    fmt.Sprintf("评论了你的文章: %s", article.Title),
				Type:       "COMMENT",
				Status:     0,
				Created:    time.Now(),
			}
			s.notifyRepo.Create(notify) // 需要在 notifyRepo 加个 Create 方法
		}
	}()

	return nil

}

// [NEW] 实现评论点赞
func (s *commentService) LikeComment(userId, commentId int) (string, error) {
	// 1. 查是否点过
	like, _ := s.repo.FindCommentLike(userId, commentId)

	if like != nil && like.Id > 0 {
		// 已点过 -> 取消
		s.repo.DeleteCommentLike(userId, commentId)
		s.repo.UpdateCommentLikesCount(commentId, -1)
		return "取消点赞", nil
	} else {
		// 未点过 -> 点赞
		newLike := &model.CommentLike{
			UserId:    userId,
			CommentId: commentId,
			Created:   time.Now(),
		}
		s.repo.AddCommentLike(newLike)
		s.repo.UpdateCommentLikesCount(commentId, 1)
		return "点赞成功", nil
	}
}
