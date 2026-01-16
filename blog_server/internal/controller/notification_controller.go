package controller
import (
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	service service.NotificationService
}

func NewNotificationController(service service.NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

// GetUnreadCount 获取未读数量
func (ctrl *NotificationController) GetUnreadCount(c *gin.Context) {
	// 暂时写死 ID=1 (管理员)，后面接了 JWT 可以从 c.Get("userId") 拿
	userId := 1 
	
	count, err := ctrl.service.GetUnreadCount(userId)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("获取消息失败"))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("data", count)) // 前端可能直接用 data
}