package controller

import (
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	notifyService service.NotificationService
}

func NewNotificationController(notifyService service.NotificationService) *NotificationController {
	return &NotificationController{notifyService: notifyService}
}

// [GET] /api/notification/unreadCount
func (ctrl *NotificationController) GetUnreadCount(c *gin.Context) {
	// 暂时写死 userId = 1，后续修复 CurrentUser 后可从 Token 获取
	userId := 1
	count, err := ctrl.notifyService.GetUnreadCount(userId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("count", count))
}

// [POST] /api/notification/getAPageNotification
func (ctrl *NotificationController) GetPage(c *gin.Context) {
	var params utils.PageParams
	if err := c.ShouldBindJSON(&params); err != nil {
		params = utils.PageParams{Page: 1, Rows: 10}
	}

	userId := 1 // 暂时写死
	res, err := ctrl.notifyService.GetPage(userId, params.Page, params.Rows)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, res)
}

// [GET] /api/notification/read/:id
func (ctrl *NotificationController) Read(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	err := ctrl.notifyService.MarkAsRead(id)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("操作失败"))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "已读"))
}

// [POST] /api/notification/readAll
func (ctrl *NotificationController) ReadAll(c *gin.Context) {
	userId := 1 // 暂时写死
	err := ctrl.notifyService.MarkAllAsRead(userId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("操作失败"))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "全部已读"))
}
