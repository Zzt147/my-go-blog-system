package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

type ReplyRepository interface {
	GetRepliesByCommentId(commentId int) ([]model.Reply, error)
	CreateReply(reply *model.Reply) error
	DeleteByCommentId(commentId int) error

	// [Like]
	FindReplyLike(userId, replyId int) (*model.ReplyLike, error)
	AddReplyLike(like *model.ReplyLike) error
	DeleteReplyLike(userId, replyId int) error
	UpdateReplyLikesCount(replyId int, step int) error
}

type replyRepository struct {
	db *gorm.DB
}

func NewReplyRepository(db *gorm.DB) ReplyRepository {
	return &replyRepository{db: db}
}

func (r *replyRepository) GetRepliesByCommentId(commentId int) ([]model.Reply, error) {
	var replies []model.Reply
	err := r.db.Where("comment_id = ?", commentId).Order("created asc").Find(&replies).Error
	return replies, err
}

func (r *replyRepository) CreateReply(reply *model.Reply) error {
	return r.db.Create(reply).Error
}

func (r *replyRepository) DeleteByCommentId(commentId int) error {
	return r.db.Where("comment_id = ?", commentId).Delete(&model.Reply{}).Error
}

// --- 点赞 ---
func (r *replyRepository) FindReplyLike(userId, replyId int) (*model.ReplyLike, error) {
	var like model.ReplyLike
	err := r.db.Where("user_id = ? AND reply_id = ?", userId, replyId).First(&like).Error
	return &like, err
}
func (r *replyRepository) AddReplyLike(like *model.ReplyLike) error {
	return r.db.Create(like).Error
}
func (r *replyRepository) DeleteReplyLike(userId, replyId int) error {
	return r.db.Where("user_id = ? AND reply_id = ?", userId, replyId).Delete(&model.ReplyLike{}).Error
}
func (r *replyRepository) UpdateReplyLikesCount(replyId int, step int) error {
	return r.db.Model(&model.Reply{}).Where("id = ?", replyId).
		UpdateColumn("likes", gorm.Expr("likes + ?", step)).Error
}
