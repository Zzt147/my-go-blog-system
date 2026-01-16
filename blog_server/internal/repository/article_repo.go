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

func (r *articleRepository) FindById(id int) (*model.Article, error) {
	var article model.Article
	result := r.db.First(&article, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &article, nil
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