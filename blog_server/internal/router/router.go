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
	// [NEW]
	commentRepo := repository.NewCommentRepository(db)
	// [NEW] 通知 Repo
	notifyRepo := repository.NewNotificationRepository(db)

	// --- Service 层 (业务逻辑) ---
	userSvc := service.NewUserService(userRepo)
	// [NEW] ArticleService 现在需要注入两个 Repo (Article + Tag)
	// 🔴 [MODIFIED] 这里必须传入 notifyRepo
	articleSvc := service.NewArticleService(articleRepo, tagRepo, notifyRepo, commentRepo)
	// [NEW] 注意这里注入了 userRepo，因为 Service 里要查用户头像
	commentSvc := service.NewCommentService(commentRepo, userRepo, notifyRepo, articleRepo)
	// [NEW] 通知 Service
	notifySvc := service.NewNotificationService(notifyRepo)

	// --- Controller 层 (接口入口) ---
	userCtrl := controller.NewUserController(userSvc)
	// [MODIFIED] ArticleController 现在需要注入 commentSvc 了！！！
	articleCtrl := controller.NewArticleController(articleSvc, commentSvc)
	fileCtrl := new(controller.FileController)
	// [NEW]
	commentCtrl := controller.NewCommentController(commentSvc)
	// [NEW] 通知 Controller
	notifyCtrl := controller.NewNotificationController(notifySvc)

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
		apiGroup.POST("/logout", userCtrl.Logout)           // 退出
		apiGroup.GET("/user/current", userCtrl.CurrentUser) // 获取当前用户
		apiGroup.GET("/users", userCtrl.ListUsers)          // 用户列表
		apiGroup.GET("/user/:id", userCtrl.GetUser)         // 用户详情

		// 用户相关
		apiGroup.GET("/user/captcha", userCtrl.Captcha)              // 图形验证码
		apiGroup.POST("/user/sendEmailCode", userCtrl.SendEmailCode) // 发送邮件验证码
		apiGroup.POST("/user/register", userCtrl.Register)           // 注册

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

		// [NEW] 二合一接口 (修复 404)
		apiGroup.POST("/article/getArticleAndFirstPageCommentByArticleId", articleCtrl.GetArticleAndFirstPageCommentByArticleId)

		// 2. 文章操作接口
		apiGroup.POST("/article/getAPageOfArticle", articleCtrl.GetPage) // 分页查询
		apiGroup.POST("/article/publishArticle", articleCtrl.Publish)    // 发布/编辑
		apiGroup.POST("/article/deleteById", articleCtrl.Delete)         // 删除

		// 3. 通用详情与列表接口
		// (这些放在最后，防止 "getAllTags" 被当成 id 解析)
		apiGroup.GET("/articles", articleCtrl.List)      // 普通列表
		apiGroup.GET("/article/:id", articleCtrl.Detail) // 文章详情

		// 🔔 通知模块
		apiGroup.GET("/notification/unreadCount", notifyCtrl.GetUnreadCount)

		// 💬 评论模块
		apiGroup.POST("/comment/getAPageCommentByArticleId", commentCtrl.GetComments)
		apiGroup.POST("/comment/insert", commentCtrl.InsertComment)

		// 🗣️ 回复模块
		apiGroup.GET("/reply/getReplies", commentCtrl.GetReplies) // 可能是 GET 或 POST
		apiGroup.POST("/reply/insert", commentCtrl.InsertReply)

		// ❤️ 点赞模块
		apiGroup.POST("/comment/likeComment", commentCtrl.LikeComment)
		// [NEW] 专门给回复用的点赞接口
		apiGroup.POST("/reply/likeReply", commentCtrl.LikeReply)
		// [NEW] 注册文章点赞接口
		apiGroup.POST("/article/likeArticle", articleCtrl.LikeArticle)
	}

	return r
}
