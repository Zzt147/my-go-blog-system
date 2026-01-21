package controller

import (
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OpLogController struct {
	opLogService service.OpLogService
}

func NewOpLogController(opLogService service.OpLogService) *OpLogController {
	return &OpLogController{opLogService: opLogService}
}

// [NEW] 获取我的足迹
// 对应 Java: @PostMapping("/opLog/getAPageOfOpLog")
// [MODIFY] 获取我的足迹 (适配 GET 请求)
// 对应路由 GET /api/oplog/getMyLogs
func (ctrl *OpLogController) GetPage(c *gin.Context) {
	var params utils.PageParams

	// [FIX] 因为前端发的是 GET，参数在 URL 里 (?page=1&rows=10)
	// BindJSON 无法处理 GET Query 参数
	if c.Request.Method == "GET" {
		pageStr := c.DefaultQuery("page", "1")
		rowsStr := c.DefaultQuery("rows", "10") // 或者是 "size"
		params.Page, _ = strconv.Atoi(pageStr)
		params.Rows, _ = strconv.Atoi(rowsStr)
		if params.Rows == 0 {
			params.Rows = 10
		}
	} else {
		c.ShouldBindJSON(&params)
	}

	userId := c.GetInt("userId")
	if userId <= 0 {
		c.JSON(http.StatusOK, utils.Error("请先登录"))
		return
	}

	res, err := ctrl.opLogService.GetMyFootprints(userId, &params)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, res)
}
