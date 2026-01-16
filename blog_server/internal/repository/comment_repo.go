package repository

import (
	"my-blog/internal/model"
	"gorm.io/gorm"
)

type CommentRepository interface {
	// 获取文章的一页评论
	GetPageByArticleId(articleId int, page int, pageSize int) ([]model.Comment, int64, error)
	Create(comment *model.Comment) error
	// 获取某条评论下的所有回复
	GetRepliesByCommentId(commentId int) ([]model.Reply, error)
	CreateReply(reply *model.Reply) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// 获取评论列表 (只获取根评论，不包含回复)
func (r *commentRepository) GetPageByArticleId(articleId int, page int, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	offset := (page - 1) * pageSize
	
	// 🔴 修复点：status = 'approved' (对应 Java/数据库逻辑)，而不是 1
	query := r.db.Model(&model.Comment{}).Where("article_id = ? AND status = ?", articleId, "approved")
	
	query.Count(&total)
	
	err := query.Order("created desc").Limit(pageSize).Offset(offset).Find(&comments).Error
	return comments, total, err
}

func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// 获取回复 (一次性把该评论下的回复都查出来)
func (r *commentRepository) GetRepliesByCommentId(commentId int) ([]model.Reply, error) {
	var replies []model.Reply
	err := r.db.Where("comment_id = ?", commentId).Order("created asc").Find(&replies).Error
	return replies, err
}

func (r *commentRepository) CreateReply(reply *model.Reply) error {
	return r.db.Create(reply).Error
}