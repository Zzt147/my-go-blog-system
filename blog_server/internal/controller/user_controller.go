package controller

import (
	//"my-blog/config"
	"my-blog/internal/model"
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	//"time"
	"strconv" // 👈 修复 undefined: strconv

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

// 验证码存储驱动 (内存模式)
var store = base64Captcha.DefaultMemStore

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{userService: userService}
}

// ListUsers 对应 GET /api/users
func (ctrl *UserController) ListUsers(c *gin.Context) {
	users, err := ctrl.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": users, "msg": "success"})
}

// GetUser 对应 GET /api/user/:id
func (ctrl *UserController) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID必须是数字"})
		return
	}

	user, err := ctrl.userService.GetUserDetail(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user, "msg": "success"})
}

// [NEW] 获取当前用户信息
// 前端刷新页面后，可能会调这个接口来维持登录状态
func (ctrl *UserController) CurrentUser(c *gin.Context) {
	// 同样，直接返回那个管理员用户
	adminUser := model.User{
		Id:       1,
		Username: "admin",
		Avatar:   "/api/images/6.png",
	}
	
	c.JSON(http.StatusOK, utils.Ok().Put("user", adminUser))
}

// [NEW] 退出登录
func (ctrl *UserController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "退出成功"))
}

// [NEW] 获取图形验证码 (/api/user/captcha)
func (ctrl *UserController) Captcha(c *gin.Context) {
	// 生成图形验证码
	driver := base64Captcha.NewDriverDigit(48, 130, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error("验证码生成失败"))
		return
	}
	
	// 这里你需要根据前端怎么传 key 来适配
	// Java 代码是前端传 key，后端存 Redis。
	// base64Captcha 库是生成 id (key) 返回给前端。
	// 为了适配你的 Java 前端：我们可能需要把 id 放在 response header 或者 body 里
	
	// 简单适配：直接把验证码的值存入 Redis，key 由前端参数决定
	key := c.Query("key") // 前端传来的随机 key
	if key != "" {
		// 注意：这里为了拿到验证码的值，base64Captcha 比较麻烦
		// 建议：直接用 verify 接口，或者这里为了简单复刻 Java 逻辑：
		// 我们把 id 当作 key 返回给前端，让前端下次带上来
		// 但你的前端逻辑已定，所以我们可以 "hack" 一下，或者让前端改用这个 id
		
		// 🛠️ 兼容策略：我们依然生成图片返回流
		// 但由于 base64Captcha 封装较深，这里直接返回 base64 字符串给前端 img src 使用
		// 前端 Java 写法是 stream 输出图片。Gin 中可以直接写流。
		// 这里简化：返回 JSON 包含 base64，前端可能需要微调，或者我们这里写死一个简单的
	}

	// ✅ 推荐做法：返回 JSON
	c.JSON(http.StatusOK, gin.H{
		"img": b64s, // Base64 图片
		"key": id,   // 验证码 ID (前端需要存这个 key 并在注册/发送邮件时传回来)
	})
}

// [NEW] 发送邮件验证码 (/api/user/sendEmailCode)
func (ctrl *UserController) SendEmailCode(c *gin.Context) {
	email := c.Query("email") // 或者是 PostForm
	// 图形验证码校验逻辑... (暂时省略，先跑通核心流程)
	
	if email == "" {
		c.JSON(http.StatusOK, utils.Error("邮箱不能为空"))
		return
	}

	err := ctrl.userService.SendEmailCode(email)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "验证码发送成功"))
}

// [NEW] 注册接口 (/api/user/register)
func (ctrl *UserController) Register(c *gin.Context) {
	// 定义一个 DTO 来接收参数
	var dto struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Code     string `json:"code"` // 验证码
	}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数错误"))
		return
	}

	user := &model.User{
		Username: dto.Username,
		Password: dto.Password,
		Email:    dto.Email,
	}

	err := ctrl.userService.Register(user, dto.Code)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Ok().Put("msg", "注册成功"))
}

// [NEW] 真实登录接口 (/api/login)
// 替代之前的假登录
func (ctrl *UserController) Login(c *gin.Context) {
	// 接收 Form 表单数据 (Spring Security 默认是 x-www-form-urlencoded)
	// 也可以兼容 JSON
	username := c.PostForm("username")
	password := c.PostForm("password")
	
	// 如果 PostForm 没拿到，试试 JSON (兼容性)
	if username == "" {
		var dto struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		c.ShouldBindJSON(&dto)
		username = dto.Username
		password = dto.Password
	}

	if username == "" || password == "" {
    c.JSON(http.StatusOK, utils.Error("用户名或密码不能为空"))
    return
  }

	user, token, err := ctrl.userService.Login(username, password)
	if err != nil {
		// 返回格式必须符合 Java 前端预期
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}

	// ---------------------------------------------------------
  // [NEW] 核心修复：手动填充权限
  // ---------------------------------------------------------
  // 模拟 Spring Security 的 GrantedAuthority 结构
  // 如果是 admin 用户，给 admin 权限；否则给 common 权限
  role := "ROLE_common"
  if user.Username == "admin" {
      role = "ROLE_admin"
  }
  
  user.Authorities = []map[string]string{
      {"authority": role}, // 结构必须是 {"authority": "ROLE_xxx"}
  }
  // ---------------------------------------------------------

	// 构造返回数据 (完全复刻 Java MyAuthenticationSuccessHandler)
	res := utils.Ok()
	res.Put("msg", "登录成功")
	res.Put("user", user) // 放入 User 对象
	res.Put("token", token) // 额外给一个 Token (虽然 Java 前端可能只用 user)
	
	// ⚠️ 关键：Java 前端可能依赖 authorites 字段
	// 你可以在 user model 里加个 Authorities []string 字段并填充
	
	c.JSON(http.StatusOK, res)
}