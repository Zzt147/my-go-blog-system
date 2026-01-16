package service

import (
	"errors"
	"my-blog/internal/model"
	"my-blog/internal/repository"
	"my-blog/pkg/utils" // 引入我们刚写的工具包
	"time"
)

// 1. 接口
type ArticleService interface {
	GetArticleList() ([]model.Article, error)
	GetArticleDetail(id int) (*model.Article, error)
	// [NEW] 对应 Java 的 getAPageOfArticle
	GetPageList(pageParams *utils.PageParams) (*utils.Result, error)

	// [NEW] 发布文章 (复刻 Java 的 publishArticle)
	Publish(article *model.Article, isEdit bool) error
	// [NEW] 删除文章
	Delete(id int) error

	// [NEW] 新增真实业务接口
	GetAllTags() ([]model.Tag, error)
	GetHotArticles() ([]model.Article, error)
	GetIndexData() (*utils.Result, error) // 聚合接口
}

// 2. 结构体
type articleService struct {
	repo repository.ArticleRepository
	tagRepo repository.TagRepository // [NEW] 引入 TagRepo
}

// 3. 构造函数
// [NEW] 修改构造函数，注入 tagRepo
func NewArticleService(repo repository.ArticleRepository, tagRepo repository.TagRepository) ArticleService {
	return &articleService{
		repo:    repo,
		tagRepo: tagRepo,
	}
}

// 4. 实现
func (s *articleService) GetArticleList() ([]model.Article, error) {
	// 这里以后可以加分页逻辑
	return s.repo.FindAll()
}

func (s *articleService) GetArticleDetail(id int) (*model.Article, error) {
	return s.repo.FindById(id)
}

// [NEW] 实现方法
func (s *articleService) GetPageList(p *utils.PageParams) (*utils.Result, error) {
	// 1. 调用 Repo 获取数据
	articles, total, err := s.repo.GetPage(p.Page, p.Rows)
	if err != nil {
		return nil, err
	}

	// 2. 组装成前端需要的 Result 格式
	// 前端通常需要 total 和 rows
	res := utils.Ok()
	res.Put("articles", articles) // 放入文章列表
	res.Put("total", total)       // 放入总数
	
	return res, nil
}

// [NEW] 实现 Publish
// 参数说明：isEdit=true 代表是编辑，false 代表是新增
func (s *articleService) Publish(article *model.Article, isEdit bool) error {
	// 1. 设置默认缩略图 (复刻 Java 逻辑)
	if article.Thumbnail == "" {
		article.Thumbnail = "/api/images/6.png" 
	}

	// 2. 自动填充时间
	now := time.Now()
	if !isEdit {
		// 如果是新增，设置创建时间
		article.Created = now
		// 暂时写死 UserID，因为还没做登录
		// 等后面做了 User 模块，这里换成从 Context 取 ID
		article.UserId = 1 
		article.Author = "Admin" 
		
		return s.repo.Create(article)
	} else {
		// 如果是编辑，设置修改时间
		article.Modified = now
		return s.repo.Update(article)
	}
}

// [NEW] 实现 Delete
func (s *articleService) Delete(id int) error {
	if id <= 0 {
		return errors.New("无效的 ID")
	}
	return s.repo.Delete(id)
}

func (s *articleService) GetHotArticles() ([]model.Article, error) {
	// 获取点赞排行 (Top 10)
	return s.repo.GetLikeRanking(10)
}

func (s *articleService) GetAllTags() ([]model.Tag, error) {
	// 获取热门标签 (Top 20)
	return s.tagRepo.GetHotTags(20)
}

// 聚合首页数据
func (s *articleService) GetIndexData() (*utils.Result, error) {
	res := utils.Ok()

	// 1. 获取标签 (Top 20)
	tags, _ := s.tagRepo.GetHotTags(20)
	// Java 代码里似乎只返回了标签名的列表？如果是，我们可以转换一下
	// 这里直接返回对象列表，前端改一下或者后端转一下都行
	// 为了兼容前端可能的 tags: ["Java", "Go"] 格式：
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}
	res.Put("tags", tagNames)      // 简单字符串数组
	res.Put("tagObjs", tags)       // 完整对象数组 (可选)

	// 2. 获取排行
	hotArticles, _ := s.repo.GetLikeRanking(10)
	res.Put("hotArticles", hotArticles) // 前端变量名通常叫 hotArticles 或 articleVOs

	// 3. 最新文章 (这里暂时只给个空列表或复用分页逻辑，根据需求)
	// latest, _, _ := s.repo.GetPage(1, 5)
	// res.Put("latestArticles", latest)

	return res, nil
}