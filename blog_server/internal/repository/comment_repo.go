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
	// [NEW] 评论点赞相关
	FindCommentLike(userId, commentId int) (*model.CommentLike, error)
	AddCommentLike(like *model.CommentLike) error
	DeleteCommentLike(userId, commentId int) error
	UpdateCommentLikesCount(commentId int, step int) error // step=1 加, step=-1 减

	// [NEW] 回复点赞相关
	FindReplyLike(userId, replyId int) (*model.ReplyLike, error)
	AddReplyLike(like *model.ReplyLike) error
	DeleteReplyLike(userId, replyId int) error
	UpdateReplyLikesCount(replyId int, step int) error
	// [NEW] 新增：分页获取评论
	// 返回值：评论列表, 总数, 错误
	GetPage(articleId, page, rows int) ([]*model.Comment, int64, error)

	// [NEW] 更新文章的评论数
	UpdateArticleCommentCount(articleId int, step int) error
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

// --- 评论点赞 ---
func (r *commentRepository) FindCommentLike(userId, commentId int) (*model.CommentLike, error) {
	var like model.CommentLike
	err := r.db.Where("user_id = ? AND comment_id = ?", userId, commentId).First(&like).Error
	return &like, err
}

func (r *commentRepository) AddCommentLike(like *model.CommentLike) error {
	return r.db.Create(like).Error
}

func (r *commentRepository) DeleteCommentLike(userId, commentId int) error {
	return r.db.Where("user_id = ? AND comment_id = ?", userId, commentId).Delete(&model.CommentLike{}).Error
}

func (r *commentRepository) UpdateCommentLikesCount(commentId int, step int) error {
	// UPDATE t_comment SET likes = likes + ? WHERE id = ?
	return r.db.Model(&model.Comment{}).Where("id = ?", commentId).
		UpdateColumn("likes", gorm.Expr("likes + ?", step)).Error
}

// --- 回复点赞 ---
func (r *commentRepository) FindReplyLike(userId, replyId int) (*model.ReplyLike, error) {
	var like model.ReplyLike
	err := r.db.Where("user_id = ? AND reply_id = ?", userId, replyId).First(&like).Error
	return &like, err
}

func (r *commentRepository) AddReplyLike(like *model.ReplyLike) error {
	return r.db.Create(like).Error
}

func (r *commentRepository) DeleteReplyLike(userId, replyId int) error {
	return r.db.Where("user_id = ? AND reply_id = ?", userId, replyId).Delete(&model.ReplyLike{}).Error
}

func (r *commentRepository) UpdateReplyLikesCount(replyId int, step int) error {
	return r.db.Model(&model.Reply{}).Where("id = ?", replyId).
		UpdateColumn("likes", gorm.Expr("likes + ?", step)).Error
}

// [NEW] 实现 GetPage
func (r *commentRepository) GetPage(articleId, page, rows int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
	var total int64

	// 1. 计算偏移量
	offset := (page - 1) * rows

	// 2. 基础查询构建器 (只查该文章的评论)
	query := r.db.Model(&model.Comment{}).Where("article_id = ?", articleId)

	// 3. 查总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 查列表 (按时间倒序：最新的在上面)
	// 这里的 Preload("User") 是假设你有关联用户表，如果没有可以去掉
	// 如果你的评论表里直接存了 avatar 和 username，那就不需要 Preload
	err := query.
		Order("created desc").
		Limit(rows).
		Offset(offset).
		Find(&comments).Error

	return comments, total, err
}

// 2. 在文件末尾实现该方法：
func (r *commentRepository) UpdateArticleCommentCount(articleId int, step int) error {
	// 逻辑：先检查统计记录是否存在，不存在则初始化，存在则更新
	var count int64
	// 注意：这里需要引入 model 包
	r.db.Table("t_statistic").Where("article_id = ?", articleId).Count(&count)

	if count == 0 {
		// 如果还没有统计记录，先创建一条 (hits=0, likes=0, comments_num=0)
		// 注意这里用 map 或者结构体插入都行，只要表名对
		r.db.Table("t_statistic").Create(map[string]interface{}{
			"article_id":   articleId,
			"comments_num": 0,
			"hits":         0,
			"likes":        0,
		})
	}

	// 执行更新：comments_num = comments_num + step
	return r.db.Table("t_statistic").
		Where("article_id = ?", articleId).
		UpdateColumn("comments_num", gorm.Expr("comments_num + ?", step)).Error
}
