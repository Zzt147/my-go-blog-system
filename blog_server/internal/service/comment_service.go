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
	GetReplies(commentId int) ([]model.Reply, error)
	AddReply(reply *model.Reply) error

	// [NEW] 点赞
	LikeComment(userId, commentId int) (string, error) // 返回 "点赞成功" 或 "取消点赞"
	LikeReply(userId, replyId int) (string, error)
}

type commentService struct {
	repo     repository.CommentRepository
	userRepo repository.UserRepository

	// [NEW] 注入通知服务和文章Repo (为了查文章作者)
	notifyRepo  repository.NotificationRepository 
	articleRepo repository.ArticleRepository
}

// [MODIFIED] 修改构造函数，注入新的依赖
func NewCommentService(
    repo repository.CommentRepository, 
    userRepo repository.UserRepository,
    notifyRepo repository.NotificationRepository, // 新增
    articleRepo repository.ArticleRepository,     // 新增
) CommentService {
	return &commentService{
        repo: repo, 
        userRepo: userRepo,
        notifyRepo: notifyRepo,
        articleRepo: articleRepo,
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

    // 2. [NEW] 发送通知 (复刻 NotificationAspect)
    go func() { // 开个协程异步发，不卡主线程
        article, _ := s.articleRepo.FindById(comment.ArticleId)
        if article != nil && article.UserId != comment.UserId { // 自己评论自己不发通知
            notify := &model.Notification{
                UserId:  article.UserId, // 接收者：文章作者
                Content: fmt.Sprintf("评论了你的文章: %s", article.Title),
                Type:    "COMMENT",
                Status:  0,
                Created: time.Now(),
            }
            s.notifyRepo.Create(notify) // 需要在 notifyRepo 加个 Create 方法
        }
    }()
    
	return nil

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

reply.Created = time.Now()
    
    // 1. 存回复
	if err := s.repo.CreateReply(reply); err != nil {
		return err
	}
    
    // 2. [NEW] 发送通知
    go func() {
        // 确定接收者：如果有 ToUid (回复某人)，发给他；否则发给层主 (暂时没办法直接查层主ID，除非再查一遍 Comment)
        receiverId := reply.ToUid
        // 如果没有指定回复谁，默认回复层主，这里为了简单先只处理 ToUid 存在的场景
        // 或者你需要再查一下 Comment 表拿到 comment.UserId
        
        if receiverId != 0 && receiverId != reply.UserId {
             notify := &model.Notification{
                UserId:  receiverId,
                Content: "回复了你的评论: " + utils.SubString(reply.Content, 20), // 截取前20字
                Type:    "REPLY",
                Status:  0,
                Created: time.Now(),
            }
            s.notifyRepo.Create(notify)
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

// [NEW] 实现回复点赞
func (s *commentService) LikeReply(userId, replyId int) (string, error) {
    like, _ := s.repo.FindReplyLike(userId, replyId)
    
    if like != nil && like.Id > 0 {
        s.repo.DeleteReplyLike(userId, replyId)
        s.repo.UpdateReplyLikesCount(replyId, -1)
        return "取消点赞", nil
    } else {
        newLike := &model.ReplyLike{
            UserId:  userId,
            ReplyId: replyId,
            Created: time.Now(),
        }
        s.repo.AddReplyLike(newLike)
        s.repo.UpdateReplyLikesCount(replyId, 1)
        return "点赞成功", nil
    }
}