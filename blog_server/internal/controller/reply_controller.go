package controller

import (
	"my-blog/internal/model"
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReplyController struct {
	replyService service.ReplyService
}

func NewReplyController(svc service.ReplyService) *ReplyController {
	return &ReplyController{replyService: svc}
}

// 严格对应: GetReplies
func (ctrl *ReplyController) GetReplies(c *gin.Context) {
	commentId, _ := strconv.Atoi(c.Query("commentId"))
	replies, err := ctrl.replyService.GetReplies(commentId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("获取回复失败"))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("replies", replies).Put("total", len(replies)))
}

// 严格对应: InsertReply (注意：接收参数是 reply)
func (ctrl *ReplyController) InsertReply(c *gin.Context) {
	var dto struct {
		CommentId int    `json:"commentId"`
		Content   string `json:"content"`
		UserId    int    `json:"userId"`
		ToUid     int    `json:"toUid"`
	}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数错误"))
		return
	}

	// [NEW] 获取真实用户ID
	realUserId := c.GetInt("userId")

	reply := model.Reply{
		CommentId: dto.CommentId,
		Content:   dto.Content,
		UserId:    realUserId, // [NEW] 使用 Token 中的 ID，而不是前端传的
		ToUid:     dto.ToUid,
		Ip:        c.ClientIP(),
		Location:  "未知",
	}

	// 调用 Service 的 AddReply
	if err := ctrl.replyService.AddReply(&reply); err != nil {
		c.JSON(http.StatusOK, utils.Error("回复失败: "+err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "回复成功"))
}

// 严格对应: LikeReply
func (ctrl *ReplyController) LikeReply(c *gin.Context) {
	replyId, _ := strconv.Atoi(c.Query("replyId"))
	// [NEW] 替换 userId := 1
	userId := c.GetInt("userId")
	msg, err := ctrl.replyService.LikeReply(userId, replyId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("操作失败"))
		return
	}
	res := utils.Ok()
	res.Msg = msg
	c.JSON(http.StatusOK, res)
}
