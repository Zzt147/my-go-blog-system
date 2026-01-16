package controller

import (
	"my-blog/pkg/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileController struct{}

// 配置路径 (对应你 application.yml 里的配置)
// 注意：Windows 路径在 Go 字符串里可以用 / 代替 \\，更不容易出错
const (
	UploadDir = "D:/my_blog_upload/"    // 对应 file.upload-avatar-dir
	UrlPrefix = "/api/file/images/"     // 前端访问的前缀
)

// Upload 处理文件上传
func (f *FileController) Upload(c *gin.Context) {
	// 1. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, utils.Error("请选择要上传的图片"))
		return
	}

	// 2. 生成新文件名 (UUID + 后缀)
	// file.Filename 是原始文件名，比如 "avatar.jpg"
	ext := filepath.Ext(file.Filename) // 获取 .jpg
	if ext == "" {
		ext = ".jpg"
	}
	// 生成类似 "a1b2-c3d4.jpg"
	newFileName := uuid.New().String() + ext

	// 3. 检查目录是否存在，不存在则创建
	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		// 0755 是读写权限设置
		os.MkdirAll(UploadDir, 0755)
	}

	// 4. 保存文件到硬盘
	// 拼接完整路径: D:/my_blog_upload/a1b2-c3d4.jpg
	destPath := filepath.Join(UploadDir, newFileName)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		c.JSON(http.StatusOK, utils.Error("文件保存失败: "+err.Error()))
		return
	}

	// 5. 返回访问 URL
	// 拼接: /api/file/images/a1b2-c3d4.jpg
	fileUrl := UrlPrefix + newFileName

	// 返回 Result 格式
	c.JSON(http.StatusOK, utils.Ok().Put("url", fileUrl).Put("msg", "上传成功"))
}