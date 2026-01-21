package controller

import (
	"my-blog/internal/model"
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	categoryService service.CategoryService
}

func NewCategoryController(categoryService service.CategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

// GET /api/category/getTree
func (ctrl *CategoryController) GetTree(c *gin.Context) {
	res, err := ctrl.categoryService.GetTree()
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, res)
}

// GET /api/category/getResources?id=
func (ctrl *CategoryController) GetResources(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.Atoi(idStr)

	res, err := ctrl.categoryService.GetResources(id)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, res)
}

// POST /api/category/add
func (ctrl *CategoryController) Add(c *gin.Context) {
	var cat model.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数错误"))
		return
	}
	if err := ctrl.categoryService.Add(&cat); err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "添加成功"))
}

// POST /api/category/update
func (ctrl *CategoryController) Update(c *gin.Context) {
	var cat model.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数错误"))
		return
	}
	if err := ctrl.categoryService.Update(&cat); err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "更新成功"))
}

// POST /api/category/updateBatch (拖拽排序)
func (ctrl *CategoryController) UpdateBatch(c *gin.Context) {
	var list []model.Category
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数错误"))
		return
	}
	if err := ctrl.categoryService.UpdateBatch(list); err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "排序已更新"))
}

// POST /api/category/delete?id=&mode=
func (ctrl *CategoryController) Delete(c *gin.Context) {
	idStr := c.Query("id")
	modeStr := c.Query("mode")
	id, _ := strconv.Atoi(idStr)
	mode, _ := strconv.Atoi(modeStr)

	// 前端默认传参可能在 query 里，也可能没有 mode (默认1)
	if mode == 0 {
		mode = 1
	}

	if err := ctrl.categoryService.Delete(id, mode); err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "删除成功"))
}
