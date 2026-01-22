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
	// 2. [MODIFY] 静态资源映射 (修复硬编码)
	// ==========================================
	// 确保 Config 已初始化
	if config.Config.File.UploadImagesDir == "" {
		config.InitConfig()
	}

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
	replyRepo := repository.NewReplyRepository(db) // [NEW] 独立
	opLogRepo := repository.NewOpLogRepository(db) // [NEW]
	// [NEW]
	categoryRepo := repository.NewCategoryRepository(db)

	// --- Service 层 (业务逻辑) ---
	// [NEW] Service (新增 MailService)
	mailSvc := service.NewMailService()
	// [MODIFY] UserService 注入 MailService
	userSvc := service.NewUserService(userRepo, mailSvc)
	// [NEW] ArticleService 现在需要注入两个 Repo (Article + Tag)
	// 🔴 [MODIFIED] 这里必须传入 notifyRepo
	//原来: articleSvc := service.NewArticleService(articleRepo, tagRepo, notifyRepo, commentRepo)
	articleSvc := service.NewArticleService(articleRepo, tagRepo, notifyRepo, commentRepo, categoryRepo)
	// [NEW] 注意这里注入了 userRepo，因为 Service 里要查用户头像
	// CommentService: 需要 ReplyRepo 用于级联删除
	commentSvc := service.NewCommentService(commentRepo, userRepo, notifyRepo, articleRepo, replyRepo)
	// [NEW] 通知 Service
	notifySvc := service.NewNotificationService(notifyRepo)
	// ReplyService: 独立
	replySvc := service.NewReplyService(replyRepo, userRepo, commentRepo, notifyRepo, articleRepo)
	opLogSvc := service.NewOpLogService(opLogRepo) // [NEW]
	// [NEW] 注入 ArticleRepo 以便级联操作文章
	categorySvc := service.NewCategoryService(categoryRepo, articleRepo)

	// --- Controller 层 (接口入口) ---
	userCtrl := controller.NewUserController(userSvc)
	// [MODIFIED] ArticleController 现在需要注入 commentSvc 了！！！
	articleCtrl := controller.NewArticleController(articleSvc, commentSvc)
	fileCtrl := new(controller.FileController)
	// [NEW]
	commentCtrl := controller.NewCommentController(commentSvc)
	// [NEW] 通知 Controller
	notifyCtrl := controller.NewNotificationController(notifySvc)
	replyCtrl := controller.NewReplyController(replySvc) // [NEW] 独立
	opLogCtrl := controller.NewOpLogController(opLogSvc) // [NEW]
	// [NEW]
	categoryCtrl := controller.NewCategoryController(categorySvc)

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
		// apiGroup.POST("/logout", userCtrl.Logout) // 退出
		// apiGroup.GET("/user/currentUser", userCtrl.CurrentUser) // 获取当前用户
		apiGroup.GET("/users", userCtrl.ListUsers)  // 用户列表
		apiGroup.GET("/user/:id", userCtrl.GetUser) // 用户详情

		// 用户相关
		apiGroup.GET("/user/captcha", userCtrl.Captcha)              // 图形验证码
		apiGroup.POST("/user/sendEmailCode", userCtrl.SendEmailCode) // 发送邮件验证码
		apiGroup.POST("/user/register", userCtrl.Register)           // 注册
		// [NEW] 重置密码
		apiGroup.POST("/user/resetPassword", userCtrl.ResetPassword)
		// ----------------------------------
		// 文件模块 (File)
		// ----------------------------------
		// apiGroup.POST("/file/upload", fileCtrl.Upload)

		// ----------------------------------
		// 文章模块 (Article)
		// ⚠️ 注意：特定路径必须放在 /article/:id 之前！
		// ----------------------------------

		// 1. 首页聚合与统计接口
		apiGroup.POST("/article/getIndexData1", articleCtrl.GetIndexData)   // 首页聚合数据 (Tags + Hot + Latest)
		apiGroup.GET("/article/getAllTags", articleCtrl.GetAllTags)         // 标签云
		apiGroup.GET("/article/getLikeRanking", articleCtrl.GetLikeRanking) // 阅读/点赞排行

		// [NEW] 阅读排行接口
		apiGroup.GET("/article/getReadRanking", articleCtrl.GetReadRanking)

		// [NEW] 文章搜索接口 (标签筛选)
		apiGroup.POST("/article/articleSearch", articleCtrl.ArticleSearch)

		// [NEW] 二合一接口 (修复 404)
		apiGroup.POST("/article/getArticleAndFirstPageCommentByArticleId", articleCtrl.GetArticleAndFirstPageCommentByArticleId)

		// 2. 文章操作接口
		apiGroup.POST("/article/getAPageOfArticle", articleCtrl.GetPage) // 分页查询
		// apiGroup.POST("/article/publishArticle", articleCtrl.Publish)    // 发布/编辑
		// apiGroup.POST("/article/deleteById", articleCtrl.Delete)         // 删除

		// 3. 通用详情与列表接口
		// (这些放在最后，防止 "getAllTags" 被当成 id 解析)
		apiGroup.GET("/articles", articleCtrl.List)      // 普通列表
		apiGroup.GET("/article/:id", articleCtrl.Detail) // 文章详情

		// 💬 Comment
		apiGroup.POST("/comment/getAPageCommentByArticleId", commentCtrl.GetComments)
		// apiGroup.POST("/comment/insert", commentCtrl.InsertComment)
		// apiGroup.POST("/comment/likeComment", commentCtrl.LikeComment)

		// 🗣️ Reply (注意：现在路由指向 replyCtrl，并且函数名严格对应 Controller 里的命名)
		apiGroup.GET("/reply/getReplies", replyCtrl.GetReplies)
		// apiGroup.POST("/reply/insert", replyCtrl.InsertReply)  // 严格对应 InsertReply
		// apiGroup.POST("/reply/likeReply", replyCtrl.LikeReply) // 严格对应 LikeReply
		// [NEW] 注册文章点赞接口
		// apiGroup.POST("/article/likeArticle", articleCtrl.LikeArticle)

		// --- [NEW] 需要登录的接口组 ---
		authGroup := apiGroup.Group("")
		authGroup.Use(middleware.Auth())
		{
			// User
			// [MODIFY] 修复路由名称，且移入 Auth 组以获取真实 UserID
			authGroup.GET("/user/currentUser", userCtrl.CurrentUser)
			authGroup.POST("/logout", userCtrl.Logout)

			// Article (写操作)
			authGroup.POST("/article/publishArticle", articleCtrl.Publish)
			authGroup.POST("/article/deleteById", articleCtrl.Delete)
			authGroup.POST("/article/likeArticle", articleCtrl.LikeArticle) // 点赞

			// File
			authGroup.POST("/file/upload", fileCtrl.Upload)

			// Comment & Reply
			authGroup.POST("/comment/insert", commentCtrl.InsertComment)
			authGroup.POST("/comment/likeComment", commentCtrl.LikeComment) // 点赞
			authGroup.POST("/reply/insert", replyCtrl.InsertReply)
			authGroup.POST("/reply/likeReply", replyCtrl.LikeReply) // 点赞

			// 🔔 通知模块
			notifyGroup := authGroup.Group("/notification")
			{
				// 获取未读数 (Top栏小红点用)
				notifyGroup.GET("/unreadCount", notifyCtrl.GetUnreadCount)

				// 获取通知列表 (消息中心用)
				notifyGroup.POST("/getAPageNotification", notifyCtrl.GetPage)

				// 标记单条已读 (点击消息时用)
				notifyGroup.GET("/read/:id", notifyCtrl.Read)

				// 标记全部已读 (一键清除)
				notifyGroup.POST("/readAll", notifyCtrl.ReadAll)
			}

			// 1. 用户个人中心操作
			authGroup.POST("/user/updateUser", userCtrl.UpdateUser)
			authGroup.POST("/user/updatePassword", userCtrl.UpdatePassword)

			// 1. 我的文章 (POST)
			// 原路径: /article/getAPageOfArticle (错) -> 修正为: /article/getMyArticles
			authGroup.POST("/article/getMyArticles", articleCtrl.GetMyArticles)

			// 2. 我点赞的文章 (POST)
			// 原路径: /article/getAPageOfMyLike (错) -> 修正为: /article/getMyLikedArticles
			authGroup.POST("/article/getMyLikedArticles", articleCtrl.GetMyLikedArticles)

			// 3. 我的评论 (POST)
			// 原路径: /comment/getAPageOfMyComment (错) -> 修正为: /comment/getMyComments
			authGroup.POST("/comment/getMyComments", commentCtrl.GetMyComments)

			// 4. 我点赞的评论 (POST)
			// 新增路径
			authGroup.POST("/comment/getMyLikedComments", commentCtrl.GetMyLikedComments)

			// 5. 我的足迹 (GET)
			// 原路径: POST /opLog/getAPageOfOpLog (错) -> 修正为: GET /oplog/getMyLogs
			// 注意：前端路径是 /oplog/... (小写oplog)，后端必须匹配
			authGroup.GET("/oplog/getMyLogs", opLogCtrl.GetPage)

			// [NEW] Category Management (分类管理)
			authGroup.GET("/category/getTree", categoryCtrl.GetTree)
			authGroup.GET("/category/getResources", categoryCtrl.GetResources)
			authGroup.POST("/category/add", categoryCtrl.Add)
			authGroup.POST("/category/update", categoryCtrl.Update)
			authGroup.POST("/category/updateBatch", categoryCtrl.UpdateBatch)
			authGroup.POST("/category/delete", categoryCtrl.Delete)
		}
	}

	return r
}
