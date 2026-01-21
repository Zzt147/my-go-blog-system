package repository

import (
	"my-blog/internal/model"

	"gorm.io/gorm"
)

// 1. 接口定义
type ArticleRepository interface {
	FindAll() ([]model.Article, error)
	FindById(id int) (*model.Article, error)
	// 以后可以在这里加 Create, Update, Delete 等方法
	// [NEW] 分页查询：返回文章列表和总条数
	// [MODIFY] 增加 sort 参数
	GetPage(page, pageSize int, sort string) ([]model.Article, int64, error)
	// [NEW] 新增方法
	Create(article *model.Article) error
	// [NEW] 更新方法
	Update(article *model.Article) error
	// [NEW] 删除方法
	Delete(id int) error
	// 获取排行 (连表查询 t_article + t_statistic)
	GetLikeRanking(limit int) ([]model.Article, error)
	// [NEW] 点赞相关
	FindArticleLike(userId, articleId int) (*model.ArticleLike, error)
	AddArticleLike(like *model.ArticleLike) error
	DeleteArticleLike(userId, articleId int) error
	UpdateArticleLikesCount(articleId int, step int) error

	// [NEW] 新增：更新阅读量
	UpdateReadCount(articleId int) error
	// [NEW] 获取阅读排行 (按 hits 倒序)
	GetReadRanking(limit int) ([]model.Article, error)

	// 需要修复的方法
	Search(page, pageSize int, condition *model.ArticleCondition) ([]model.Article, int64, error)
	GetMyLikedArticles(userId, page, pageSize int) ([]model.Article, int64, error)

	// [NEW] 根据分类ID查询文章 (用于 getResources)
	FindByCategoryId(categoryId int) ([]model.Article, error)

	// [NEW] 批量更新文章的分类 (用于删除分类模式1: 移动文章到父级)
	UpdateCategoryId(oldCategoryId, newCategoryId int) error

	// [NEW] 根据分类ID删除文章 (用于删除分类模式2: 销毁文章)
	DeleteByCategoryId(categoryId int) error
}

// 2. 结构体实现
type articleRepository struct {
	db *gorm.DB
}

// 3. 构造函数
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

// 4. 具体实现
func (r *articleRepository) FindAll() ([]model.Article, error) {
	var articles []model.Article
	// 相当于 select * from t_article order by created desc
	result := r.db.Order("created desc").Find(&articles)
	return articles, result.Error
}

// [MODIFIED] 1. 修复 FindById：联表查询统计数据
func (r *articleRepository) FindById(id int) (*model.Article, error) {
	var article model.Article
	// 核心 SQL: SELECT t_article.*, s.likes, s.hits AS views FROM t_article LEFT JOIN t_statistic s ON ...
	err := r.db.Table("t_article").
		Select("t_article.*, IFNULL(s.likes, 0) as likes, IFNULL(s.hits, 0) as views").
		Joins("LEFT JOIN t_statistic s ON s.article_id = t_article.id").
		Where("t_article.id = ?", id).
		First(&article).Error
	return &article, err
}

// [NEW] 实现分页查询
func (r *articleRepository) GetPage(page, pageSize int, sort string) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	// 1. 计算偏移量
	offset := (page - 1) * pageSize

	// 使用 Table + Join 以便获取 hits 和 likes，并支持按 hits 排序
	query := r.db.Table("t_article").
		Select("t_article.*, t_statistic.likes, t_statistic.hits AS views").
		Joins("LEFT JOIN t_statistic ON t_article.id = t_statistic.article_id")

	// 处理排序逻辑
	if sort == "hot" {
		// [NEW] 按热度(阅读量)倒序
		query = query.Order("t_statistic.hits DESC")
	} else {
		// [MODIFY] 默认按时间倒序
		query = query.Order("t_article.created DESC")
	}

	// 3. 先查总数 (Count)
	// 对应 Java MyBatis Plus 的 selectCount
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 再查列表 (Limit Offset)
	// 对应 SQL: SELECT * FROM t_article ORDER BY created DESC LIMIT 10 OFFSET 0
	// 查询列表
	err := query.Limit(pageSize).Offset(offset).Find(&articles).Error
	return articles, total, err
}

// [NEW] 实现 Create
func (r *articleRepository) Create(article *model.Article) error {
	// GORM 的 Create 方法会自动把结构体字段映射到 SQL 的 INSERT 语句
	// 如果 article.Id 是自增，插入后 GORM 会自动把生成的 ID 填回 article.Id
	return r.db.Create(article).Error
}

// [NEW] 实现 Update
func (r *articleRepository) Update(article *model.Article) error {
	// Model(&model.Article{}) 指定要操作的表
	// Where("id = ?", ...) 指定要更新哪一行
	// Updates(article) 会更新所有非零值字段
	// ⚠️ 注意：如果你的 int 字段值为 0，GORM 默认认为你不更新它。
	// 但在这个场景下通常没问题，因为文章ID肯定不为0。
	return r.db.Model(&model.Article{}).Where("id = ?", article.Id).Updates(article).Error
}

// [NEW] 实现 Delete
func (r *articleRepository) Delete(id int) error {
	// 对应 SQL: DELETE FROM t_article WHERE id = ?
	return r.db.Delete(&model.Article{}, id).Error
}

// GetLikeRanking 获取点赞排行
// Java逻辑: select * from t_article a left join t_statistic s on a.id = s.article_id order by s.likes desc
func (r *articleRepository) GetLikeRanking(limit int) ([]model.Article, error) {
	var articles []model.Article

	// GORM 连表查询
	// Select: 把 t_article 的字段选出来，顺便把 t_statistic 的 likes 选出来并起个别名
	err := r.db.Table("t_article").
		Select("t_article.*, t_statistic.likes, t_statistic.hits").
		Joins("LEFT JOIN t_statistic ON t_article.id = t_statistic.article_id").
		Order("t_statistic.likes DESC"). // 按点赞倒序
		Limit(limit).
		Scan(&articles).Error // Scan 会自动把查出来的 likes 填入 Article 结构体的 Likes 字段(因为字段名匹配)

	return articles, err
}

// 👇👇👇 追加在文件末尾 👇👇👇

func (r *articleRepository) FindArticleLike(userId, articleId int) (*model.ArticleLike, error) {
	var like model.ArticleLike
	err := r.db.Where("user_id = ? AND article_id = ?", userId, articleId).First(&like).Error
	return &like, err
}

func (r *articleRepository) AddArticleLike(like *model.ArticleLike) error {
	return r.db.Create(like).Error
}

func (r *articleRepository) DeleteArticleLike(userId, articleId int) error {
	return r.db.Where("user_id = ? AND article_id = ?", userId, articleId).Delete(&model.ArticleLike{}).Error
}

// [MODIFIED] 2. 修复 UpdateArticleLikesCount：更新 t_statistic 表
func (r *articleRepository) UpdateArticleLikesCount(articleId int, step int) error {
	// 先检查统计记录是否存在，不存在则创建（防止报错）
	var count int64
	r.db.Model(&model.Statistic{}).Where("article_id = ?", articleId).Count(&count)
	if count == 0 {
		r.db.Create(&model.Statistic{ArticleId: articleId, Likes: 0, Hits: 0})
	}

	// 更新 likes 字段
	return r.db.Model(&model.Statistic{}).
		Where("article_id = ?", articleId).
		UpdateColumn("likes", gorm.Expr("likes + ?", step)).Error
}

// [NEW] 实现 UpdateReadCount
func (r *articleRepository) UpdateReadCount(articleId int) error {
	// 1. 先检查统计记录是否存在
	var count int64
	r.db.Model(&model.Statistic{}).Where("article_id = ?", articleId).Count(&count)

	// 2. 如果不存在（比如新文章），先创建一条
	if count == 0 {
		// 默认 likes=0, hits=1
		r.db.Create(&model.Statistic{ArticleId: articleId, Likes: 0, Hits: 1})
		return nil
	}

	// 3. 存在则直接 +1
	return r.db.Model(&model.Statistic{}).
		Where("article_id = ?", articleId).
		UpdateColumn("hits", gorm.Expr("hits + ?", 1)).Error
}

// [NEW] 获取阅读排行实现
// Java逻辑: select ... order by s.hits desc
// [MODIFY] 获取阅读排行实现 (修复 views 为 0 的问题)
func (r *articleRepository) GetReadRanking(limit int) ([]model.Article, error) {
	var articles []model.Article

	// 关键修改：t_statistic.hits AS views
	// 这样 GORM 才能把 hits 的值赋给结构体里的 Views 字段
	err := r.db.Table("t_article").
		Select("t_article.*, t_statistic.likes, t_statistic.hits AS views").
		Joins("LEFT JOIN t_statistic ON t_article.id = t_statistic.article_id").
		Order("t_statistic.hits DESC").
		Limit(limit).
		Scan(&articles).Error

	return articles, err
}

// [NEW] 实现获取我点赞的文章
// 逻辑：联表 t_article 和 t_article_like
// [MODIFY] 修复 GetMyLikedArticles
func (r *articleRepository) GetMyLikedArticles(userId, page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// 修正：使用 Model 自动映射
	query := r.db.Model(&model.Article{}).
		Select("t_article.*, t_statistic.likes, t_statistic.hits AS views").
		Joins("JOIN t_article_like ON t_article_like.article_id = t_article.id").
		Joins("LEFT JOIN t_statistic ON t_article.id = t_statistic.article_id").
		Where("t_article_like.user_id = ?", userId).
		Order("t_article_like.created desc")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(pageSize).Offset(offset).Find(&articles).Error
	return articles, total, err
}

// [MODIFY] 修复 Search / SearchV2 (对应 Controller 里的 GetMyArticles)
// 请确保 Controller 调用的是这个 repo 方法
func (r *articleRepository) Search(page, pageSize int, condition *model.ArticleCondition) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	query := r.db.Table("t_article").
		Select("t_article.*, t_statistic.likes, t_statistic.hits AS views").
		Joins("LEFT JOIN t_statistic ON t_article.id = t_statistic.article_id").
		Order("t_article.created desc")

	if condition != nil {
		if condition.Tag != "" {
			query = query.Where("t_article.tags LIKE ?", "%"+condition.Tag+"%")
		}
		if condition.UserId > 0 {
			query = query.Where("t_article.user_id = ?", condition.UserId)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(pageSize).Offset(offset).Find(&articles).Error
	return articles, total, err
}

// [NEW] 根据分类ID查询
func (r *articleRepository) FindByCategoryId(categoryId int) ([]model.Article, error) {
	var articles []model.Article
	// 这里假设 t_article 表中有 category_id 字段 (如果没有请在数据库添加)
	// 对应 Go Model 中的 CategoryId
	err := r.db.Where("category_id = ?", categoryId).Order("created desc").Find(&articles).Error
	return articles, err
}

// [NEW] 移动文章分类
func (r *articleRepository) UpdateCategoryId(oldCategoryId, newCategoryId int) error {
	return r.db.Model(&model.Article{}).
		Where("category_id = ?", oldCategoryId).
		Update("category_id", newCategoryId).Error
}

// [NEW] 删除某分类下的所有文章
func (r *articleRepository) DeleteByCategoryId(categoryId int) error {
	return r.db.Where("category_id = ?", categoryId).Delete(&model.Article{}).Error
}
