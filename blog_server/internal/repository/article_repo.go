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
	GetPage(page int, pageSize int) ([]model.Article, int64, error)
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

	// [NEW] 文章搜索 (支持按标签筛选)
	Search(page, pageSize int, condition *model.ArticleCondition) ([]model.Article, int64, error)
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
func (r *articleRepository) GetPage(page int, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	// 1. 计算偏移量
	offset := (page - 1) * pageSize

	// 2. 构造基础查询（按时间倒序）
	query := r.db.Model(&model.Article{}).Order("created desc")

	// 3. 先查总数 (Count)
	// 对应 Java MyBatis Plus 的 selectCount
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. 再查列表 (Limit Offset)
	// 对应 SQL: SELECT * FROM t_article ORDER BY created DESC LIMIT 10 OFFSET 0
	result := query.Limit(pageSize).Offset(offset).Find(&articles)

	return articles, total, result.Error
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

// [NEW] 文章搜索实现
func (r *articleRepository) Search(page, pageSize int, condition *model.ArticleCondition) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	offset := (page - 1) * pageSize

	// 构造基础查询
	// 关键修改1：使用 Table 而不是 Model，以便进行 Join 操作
	// 关键修改2：Select 中增加 AS views，确保列表页也能显示阅读量
	query := r.db.Table("t_article").
		Select("t_article.*, t_statistic.likes, t_statistic.hits AS views").
		Joins("LEFT JOIN t_statistic ON t_article.id = t_statistic.article_id").
		Order("t_article.created desc")

	// 动态 SQL (对应 Java queryWrapper)
	if condition != nil {
		if condition.Tag != "" {
			// 注意：这里需要明确指定表名 t_article.tags，避免字段歧义（如果有的话）
			query = query.Where("t_article.tags LIKE ?", "%"+condition.Tag+"%")
		}
		// 这里预留了 CategoryId, Title 等扩展位置，严格遵循 Java ArticleCondition
	}

	// 查总数
	// 查总数 (Count 时尽量移除 Select 和 Order 以提高性能，但 GORM 的 Count 会自动处理)
	// 注意：对于连表查询的 Count，GORM 通常能处理，但为了稳健，直接对 query 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查列表
	result := query.Limit(pageSize).Offset(offset).Find(&articles)
	return articles, total, result.Error
}
