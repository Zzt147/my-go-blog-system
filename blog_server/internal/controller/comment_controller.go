package controller

import (
	"my-blog/internal/model"
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	commentService service.CommentService
}

func NewCommentController(svc service.CommentService) *CommentController {
	return &CommentController{commentService: svc}
}

// [POST] 获取文章评论
func (ctrl *CommentController) GetComments(c *gin.Context) {
	articleIdStr := c.Query("articleId")
	articleId, _ := strconv.Atoi(articleIdStr)

	var pageParams utils.PageParams
	if err := c.ShouldBindJSON(&pageParams); err != nil {
		pageParams = utils.PageParams{Page: 1, Rows: 10}
	}

	result, err := ctrl.commentService.GetComments(articleId, &pageParams)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, result)
}

// [POST] 发表评论 /api/comment/insert
func (ctrl *CommentController) InsertComment(c *gin.Context) {
	// 1. 定义 DTO 接收前端参数 (ArticleAndComment.vue 传的是 author 和 articleId)
	var dto struct {
		ArticleId interface{} `json:"articleId"` // 容错：接收 string 或 int
		Content   string      `json:"content"`
		Author    string      `json:"author"`    // 前端传过来的用户名
	}

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数解析失败: "+err.Error()))
		return
	}

	// 2. 处理 ArticleId
	var finalArticleId int
	switch v := dto.ArticleId.(type) {
	case float64:
		finalArticleId = int(v)
	case string:
		finalArticleId, _ = strconv.Atoi(v)
	case int:
		finalArticleId = v
	}

	// 3. 构造 Model
	comment := model.Comment{
		ArticleId: finalArticleId,
		Content:   dto.Content,
		Author:    dto.Author,     // 先存前端传的 author
		Ip:        c.ClientIP(),   // 获取真实 IP
		Location:  "未知",           // 后面可以用 ip2region 库解析
	}

	// 4. 调用 Service (它会自动根据 Author 查 UserId)
	if err := ctrl.commentService.AddComment(&comment); err != nil {
		c.JSON(http.StatusOK, utils.Error("评论失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Ok().Put("msg", "评论成功").Put("Comment", comment))
}

// [GET] 获取回复
func (ctrl *CommentController) GetReplies(c *gin.Context) {
	commentIdStr := c.Query("commentId")
	commentId, _ := strconv.Atoi(commentIdStr)
	
	replies, err := ctrl.commentService.GetReplies(commentId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("获取回复失败"))
		return
	}
	// 兼容前端: res.data.map.replies
	c.JSON(http.StatusOK, utils.Ok().Put("replies", replies).Put("total", len(replies)))
}

// [POST] 发表回复 /api/reply/insert
func (ctrl *CommentController) InsertReply(c *gin.Context) {
	// 1. 定义 DTO (Comment.vue 传的是 userId 和 toUid)
	var dto struct {
		CommentId int    `json:"commentId"`
		Content   string `json:"content"`
		UserId    int    `json:"userId"` // 前端传的是 ID
		ToUid     int    `json:"toUid"`  // 目标用户 ID
	}

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数解析失败"))
		return
	}

	// 2. 构造 Reply
	reply := model.Reply{
		CommentId: dto.CommentId,
		Content:   dto.Content,
		UserId:    dto.UserId,
		ToUid:     dto.ToUid,
		Ip:        c.ClientIP(),
		Location:  "未知",
	}

	// 3. 调用 Service (它会自动补全 Author 名字)
	if err := ctrl.commentService.AddReply(&reply); err != nil {
		c.JSON(http.StatusOK, utils.Error("回复失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Ok().Put("msg", "回复成功"))
}

// ... 之前的代码 ...

// [POST] 评论点赞 /api/comment/likeComment
func (ctrl *CommentController) LikeComment(c *gin.Context) {
    commentIdStr := c.Query("commentId")
    commentId, _ := strconv.Atoi(commentIdStr)
    
    // 获取当前用户ID (暂时写死1，或者从 Context 拿)
    userId := 1 
    
    msg, err := ctrl.commentService.LikeComment(userId, commentId)
    if err != nil {
        c.JSON(http.StatusOK, utils.Error("操作失败"))
        return
    }
    
		// [FIX] 直接修改 Msg 字段，而不是用 Put
		res := utils.Ok()
		res.Msg = msg // 这样前端 res.data.msg 就能读到 "点赞成功" 或 "取消点赞"
		c.JSON(http.StatusOK, res)
}

// [POST] 回复点赞 /api/reply/like
func (ctrl *CommentController) LikeReply(c *gin.Context) {
    replyIdStr := c.Query("replyId")
    replyId, _ := strconv.Atoi(replyIdStr)
    
    userId := 1 // 暂时写死
    
    msg, err := ctrl.commentService.LikeReply(userId, replyId)
    if err != nil {
        c.JSON(http.StatusOK, utils.Error("操作失败"))
        return
    }

		// [FIX] 同上，直接修改 Msg
		res := utils.Ok()
		res.Msg = msg
		c.JSON(http.StatusOK, res)
}