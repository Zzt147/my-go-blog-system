package router

import (
	"my-blog/config"
	"my-blog/internal/controller"
	"my-blog/internal/middleware"
	"my-blog/internal/repository"
	"my-blog/internal/service"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// ==========================================
	// 1. 全局中间件
	// ==========================================
	// 允许跨域请求 (Vue 端口访问 Go 端口)
	r.Use(middleware.Cors())

	// ==========================================
	// 2. 静态资源映射 (对应 Java WebConfig)
	// ==========================================
	// 头像/上传文件夹
	r.Static("/api/file/images", "D:/my_blog_upload")
	// 系统图片文件夹
	r.Static("/api/images", "E:/img/images")
	// 文章图片文件夹
	r.Static("/api/article_img", "E:/img/article_img")

	// ==========================================
	// 3. 依赖注入 (DI) 层层组装
	// ==========================================
	db := config.DB

	// 初始化 Redis (这一步别忘了！)
	config.InitRedis()

	// --- Repository 层 (数据访问) ---
	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	tagRepo := repository.NewTagRepository(db) // [NEW] 新增 TagRepo

	// --- Service 层 (业务逻辑) ---
	userSvc := service.NewUserService(userRepo)
	// [NEW] ArticleService 现在需要注入两个 Repo (Article + Tag)
	articleSvc := service.NewArticleService(articleRepo, tagRepo)

	// --- Controller 层 (接口入口) ---
	userCtrl := controller.NewUserController(userSvc)
	articleCtrl := controller.NewArticleController(articleSvc)
	fileCtrl := new(controller.FileController)

	// ==========================================
	// 4. 路由注册
	// ==========================================
	apiGroup := r.Group("/api")
	{
		// ----------------------------------
		// 用户模块 (User)
		// ----------------------------------
		// 登录 (替换原来的假登录)
		// 注意：Spring Security 默认拦截 /api/login，所以这里必须匹配
		apiGroup.POST("/login", userCtrl.Login)
		apiGroup.POST("/logout", userCtrl.Logout)        // 退出
		apiGroup.GET("/user/current", userCtrl.CurrentUser)  // 获取当前用户
		apiGroup.GET("/users", userCtrl.ListUsers)           // 用户列表
		apiGroup.GET("/user/:id", userCtrl.GetUser)          // 用户详情

		// 用户相关
		apiGroup.GET("/user/captcha", userCtrl.Captcha)       // 图形验证码
		apiGroup.POST("/user/sendEmailCode", userCtrl.SendEmailCode) // 发送邮件验证码
		apiGroup.POST("/user/register", userCtrl.Register)    // 注册

		// ----------------------------------
		// 文件模块 (File)
		// ----------------------------------
		apiGroup.POST("/file/upload", fileCtrl.Upload)

		// ----------------------------------
		// 文章模块 (Article)
		// ⚠️ 注意：特定路径必须放在 /article/:id 之前！
		// ----------------------------------
		
		// 1. 首页聚合与统计接口
		apiGroup.POST("/article/getIndexData1", articleCtrl.GetIndexData)   // 首页聚合数据 (Tags + Hot + Latest)
		apiGroup.GET("/article/getAllTags", articleCtrl.GetAllTags)         // 标签云
		apiGroup.GET("/article/getLikeRanking", articleCtrl.GetLikeRanking) // 阅读/点赞排行

		// 2. 文章操作接口
		apiGroup.POST("/article/getAPageOfArticle", articleCtrl.GetPage) // 分页查询
		apiGroup.POST("/article/publishArticle", articleCtrl.Publish)    // 发布/编辑
		apiGroup.POST("/article/deleteById", articleCtrl.Delete)         // 删除

		// 3. 通用详情与列表接口
		// (这些放在最后，防止 "getAllTags" 被当成 id 解析)
		apiGroup.GET("/articles", articleCtrl.List)    // 普通列表
		apiGroup.GET("/article/:id", articleCtrl.Detail) // 文章详情
	}

	return r
}