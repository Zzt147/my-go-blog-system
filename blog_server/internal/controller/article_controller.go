package controller

import (
	"my-blog/internal/model"
	"my-blog/internal/service"
	"my-blog/pkg/utils" // 引入我们刚写的工具包
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleController struct {
	articleService service.ArticleService
	// [NEW] 注入 CommentService，为了实现“二合一”接口
	commentService service.CommentService
}

func NewArticleController(articleService service.ArticleService, commentService service.CommentService) *ArticleController {
	return &ArticleController{articleService: articleService, commentService: commentService}
}

// 修改 List 方法 (演示旧接口怎么改用 Result)
func (ctrl *ArticleController) List(c *gin.Context) {
	articles, err := ctrl.articleService.GetArticleList()
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	// 链式调用，非常优雅
	c.JSON(http.StatusOK, utils.Ok().Put("data", articles))
}

// Detail 对应 GET /api/article/:id
// 修改 Detail 方法
func (ctrl *ArticleController) Detail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	article, err := ctrl.articleService.GetArticleDetail(id)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("文章不存在"))
		return
	}

	// 这里 key 用 "article" 对应前端
	c.JSON(http.StatusOK, utils.Ok().Put("article", article))
}

// [NEW] 对应 Java 的 @PostMapping("/getAPageOfArticle")
func (ctrl *ArticleController) GetPage(c *gin.Context) {
	// 1. 接收前端传来的 JSON 参数
	var pageParams utils.PageParams
	// ShouldBindJSON 会自动把 request body 里的 {"page":1, "rows":10} 映射到结构体
	if err := c.ShouldBindJSON(&pageParams); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数格式错误"))
		return
	}

	// 2. 调用 Service
	result, err := ctrl.articleService.GetPageList(&pageParams)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("查询失败: "+err.Error()))
		return
	}

	// 3. 直接返回 Result 结构体
	c.JSON(http.StatusOK, result)
}

// [NEW] 发布文章接口
// 对应 Java: @RequestMapping("/publishArticle")
func (ctrl *ArticleController) Publish(c *gin.Context) {
	// 1. 定义一个临时结构体，用来接收前端参数
	// 前端不仅传了 article 数据， URL 上可能还带了 ?type=add
	// 但通常 POST 请求里，type 也会放在 JSON 里，或者我们根据 ID 是否存在来判断

	var article model.Article
	// 2. 将 JSON 绑定到 article 结构体
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数格式错误: "+err.Error()))
		return
	}

	// 3. 获取 query 参数 type (例如 /publishArticle?type=add)
	actionType := c.Query("type")

	// 4. 判断是新增还是编辑
	isEdit := false
	if actionType == "edit" || article.Id > 0 {
		isEdit = true
	}

	// 5. 调用 Service
	err := ctrl.articleService.Publish(&article, isEdit)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("操作失败: "+err.Error()))
		return
	}

	res := utils.Ok()
	res.Msg = "操作成功！"
	c.JSON(http.StatusOK, res)
}

// [NEW] 删除文章接口
// 对应 Java: @PostMapping("/deleteById")
func (ctrl *ArticleController) Delete(c *gin.Context) {
	// 前端可能是 form-data 传 id，也可能是 query param
	// 假设是 Form 表单 (POST) 里的 id
	idStr := c.PostForm("id")
	if idStr == "" {
		// 如果 PostForm 没取到，试试 query param (兼容性处理)
		idStr = c.Query("id")
	}

	id, _ := strconv.Atoi(idStr)

	if err := ctrl.articleService.Delete(id); err != nil {
		c.JSON(http.StatusOK, utils.Error("删除失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Ok().Put("msg", "删除成功"))
}

// ... 之前的代码 ...
// [REAL] 获取所有标签
func (ctrl *ArticleController) GetAllTags(c *gin.Context) {
	tags, err := ctrl.articleService.GetAllTags() // 这里拿到的是 []Tag 对象
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("获取标签失败"))
		return
	}

	// 兼容 Java 逻辑：前端可能只需要由名字组成的数组
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	res := utils.Ok()
	res.Put("tags", tagNames) // 对应 Java: result.getMap().put("tags", tagNames);
	res.Put("tagObjs", tags)  // 顺便把带数量的也返回去

	c.JSON(http.StatusOK, res)
}

// [REAL] 获取排行
func (ctrl *ArticleController) GetLikeRanking(c *gin.Context) {
	articles, err := ctrl.articleService.GetHotArticles()
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("获取排行失败"))
		return
	}
	// Java 代码里这个 key 叫 "articleVOs"
	c.JSON(http.StatusOK, utils.Ok().Put("articleVOs", articles))
}

// [REAL] 首页聚合接口
func (ctrl *ArticleController) GetIndexData(c *gin.Context) {
	// 调用 Service 的聚合方法
	result, _ := ctrl.articleService.GetIndexData()
	c.JSON(http.StatusOK, result)
}

// [NEW] 二合一接口：获取文章详情 + 第一页评论
// 对应 Java: getArticleAndFirstPageCommentByArticleId
// [NEW] 对应前端的 /getArticleAndFirstPageCommentByArticleId
func (ctrl *ArticleController) GetArticleAndFirstPageCommentByArticleId(c *gin.Context) {
	// 1. 获取 Query 里的 articleId
	articleIdStr := c.Query("articleId")
	articleId, _ := strconv.Atoi(articleIdStr)

	// 2. 获取 Body 里的参数 (其实 Service 层暂时写死了取第一页，但为了兼容前端传参，我们还是读一下)
	var params utils.PageParams
	c.ShouldBindJSON(&params)

	// 3. [核心] 获取当前登录用户 ID
	// 如果还没接 JWT，暂时写死 1。
	// 如果接了，用 userId := c.GetInt("userId")
	userId := 1

	// 4. 调用我们在 Service 层写好的“超级接口”
	// 这个接口会同时搞定：文章详情 + 是否点赞(IsLiked) + 第一页评论
	res, err := ctrl.articleService.GetArticleAndFirstPageCommentByArticleId(articleId, userId)

	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, res)
}

// [POST] 文章点赞 /api/article/likeArticle
func (ctrl *ArticleController) LikeArticle(c *gin.Context) {
	// 获取 articleId
	articleIdStr := c.Query("articleId")
	articleId, _ := strconv.Atoi(articleIdStr)

	// 获取 userId (暂时写死，后续接 JWT)
	userId := 1

	msg, err := ctrl.articleService.LikeArticle(userId, articleId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("操作失败"))
		return
	}

	// [IMPORTANT] 记得像评论点赞一样，直接覆盖 Msg
	res := utils.Ok()
	res.Msg = msg
	c.JSON(http.StatusOK, res)
}

// [NEW] 获取阅读排行接口
// 对应 Java: @GetMapping("/getReadRanking")
func (ctrl *ArticleController) GetReadRanking(c *gin.Context) {
	articles, err := ctrl.articleService.GetReadRanking()
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("获取阅读排行失败"))
		return
	}
	// Java 代码返回 key 为 "articleVOs"
	c.JSON(http.StatusOK, utils.Ok().Put("articleVOs", articles))
}

// [NEW] 文章搜索接口 (按标签)
// 对应 Java: @PostMapping("/articleSearch")
func (ctrl *ArticleController) ArticleSearch(c *gin.Context) {
	// 定义请求参数结构体，匹配前端 JSON 结构
	// 前端传参: { "pageParams": {...}, "articleCondition": {...} }
	var req struct {
		PageParams       utils.PageParams       `json:"pageParams"`
		ArticleCondition model.ArticleCondition `json:"articleCondition"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数格式错误"))
		return
	}

	result, err := ctrl.articleService.Search(&req.PageParams, &req.ArticleCondition)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("查询失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}
