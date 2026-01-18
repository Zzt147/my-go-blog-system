package service

import (
	"errors"
	"fmt"
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
	// [NEW] 文章点赞
	LikeArticle(userId, articleId int) (string, error)
	// [NEW] 核心修复：聚合接口（文章详情 + 点赞状态 + 第一页评论）
	GetArticleAndFirstPageCommentByArticleId(articleId, userId int) (*utils.Result, error)

	// [NEW] 获取阅读排行
	GetReadRanking() ([]model.Article, error)

	// [NEW] 文章搜索
	Search(pageParams *utils.PageParams, articleCondition *model.ArticleCondition) (*utils.Result, error)
}

// 2. 结构体
type articleService struct {
	repo    repository.ArticleRepository
	tagRepo repository.TagRepository // [NEW] 引入 TagRepo
	// [NEW] 注入通知 Repo
	notifyRepo repository.NotificationRepository
	// [NEW] 新增：为了获取评论列表，需要注入评论仓库
	commentRepo repository.CommentRepository
}

// 3. 构造函数
// [NEW] 修改构造函数，注入 tagRepo
func NewArticleService(
	repo repository.ArticleRepository,
	tagRepo repository.TagRepository,
	notifyRepo repository.NotificationRepository,
	commentRepo repository.CommentRepository, // 新增参数
) ArticleService {
	return &articleService{
		repo:        repo,
		tagRepo:     tagRepo,
		notifyRepo:  notifyRepo,
		commentRepo: commentRepo,
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
	// 🔴 [新增校验] 必须要有标题
	if article.Title == "" {
		return errors.New("文章标题不能为空")
	}
	// (可选) 如果你也想校验内容，可以把下面这行解开
	if article.Content == "" {
		return errors.New("文章内容不能为空")
	}

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
		article.Modified = &now
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
	res.Put("tags", tagNames) // 简单字符串数组
	res.Put("tagObjs", tags)  // 完整对象数组 (可选)

	// 2. 获取排行
	hotArticles, _ := s.repo.GetLikeRanking(10)
	res.Put("hotArticles", hotArticles) // 前端变量名通常叫 hotArticles 或 articleVOs

	// 3. 最新文章 (这里暂时只给个空列表或复用分页逻辑，根据需求)
	// latest, _, _ := s.repo.GetPage(1, 5)
	// res.Put("latestArticles", latest)

	return res, nil
}

// [核心修复] 获取文章详情及相关数据
func (s *articleService) GetArticleAndFirstPageCommentByArticleId(articleId, userId int) (*utils.Result, error) {
	// 1. 查文章
	article, err := s.repo.FindById(articleId)
	if err != nil {
		return nil, err
	}

	// 2. 增加阅读数 (Hits)
	// (确保你的 article_repo.go 里有 UpdateReadCount 方法)
	s.repo.UpdateReadCount(articleId)

	// 3. [关键] 填充 IsLiked 状态
	// 如果用户登录了 (userId > 0)，去查点赞表
	if userId > 0 {
		like, _ := s.repo.FindArticleLike(userId, articleId)
		if like != nil && like.Id > 0 {
			article.IsLiked = true
		} else {
			article.IsLiked = false
		}
	}

	// 4. 查第一页评论 (默认取 5 条，按最新排序)
	// 这里调用了新注入的 commentRepo
	comments, total, _ := s.commentRepo.GetPage(articleId, 1, 5)

	// 5. 组装结果
	res := utils.Ok()
	res.Put("article", article)
	res.Put("comments", comments)
	res.Put("total", total)

	return res, nil
}

// 👇👇👇 追加 LikeArticle 实现 👇👇👇

func (s *articleService) LikeArticle(userId, articleId int) (string, error) {
	// 1. 查是否点过
	like, _ := s.repo.FindArticleLike(userId, articleId)

	if like != nil && like.Id > 0 {
		// --- 取消点赞 ---
		s.repo.DeleteArticleLike(userId, articleId)
		s.repo.UpdateArticleLikesCount(articleId, -1)
		return "取消点赞", nil
	} else {
		// --- 新增点赞 ---
		newLike := &model.ArticleLike{
			UserId:    userId,
			ArticleId: articleId,
			Created:   time.Now(),
		}
		if err := s.repo.AddArticleLike(newLike); err != nil {
			return "", err
		}
		s.repo.UpdateArticleLikesCount(articleId, 1)

		// --- 发送通知 ---
		go func() {
			// 查文章作者
			article, _ := s.repo.FindById(articleId)
			if article != nil && article.UserId != userId {
				notify := &model.Notification{
					ReceiverId: article.UserId,
					Content:    fmt.Sprintf("点赞了你的文章: %s", article.Title),
					Type:       "LIKE", // 通知类型
					Status:     0,
					Created:    time.Now(),
				}
				s.notifyRepo.Create(notify)
			}
		}()

		return "点赞成功", nil
	}
}

// [NEW] 实现 GetReadRanking
func (s *articleService) GetReadRanking() ([]model.Article, error) {
	// 获取阅读排行 (Top 10)
	return s.repo.GetReadRanking(10)
}

// [NEW] 实现 Search (对应 Java 的 search 方法)
func (s *articleService) Search(p *utils.PageParams, condition *model.ArticleCondition) (*utils.Result, error) {
	// 调用 Repo 进行搜索
	articles, total, err := s.repo.Search(p.Page, p.Rows, condition)
	if err != nil {
		return nil, err
	}

	res := utils.Ok()
	res.Put("articles", articles)
	res.Put("total", total)
	return res, nil
}
