package controller

import (
	"my-blog/config"
	"time"

	//"my-blog/config"
	"my-blog/internal/model"
	"my-blog/internal/service"
	"my-blog/pkg/utils"
	"net/http"
	//"time"
	"strconv" // 👈 修复 undefined: strconv
	"strings" // [NEW] 用于转小写

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
// [MODIFY] 获取图形验证码 (/api/user/captcha)
func (ctrl *UserController) Captcha(c *gin.Context) {
	// Java逻辑：SpecCaptcha specCaptcha = new SpecCaptcha(130, 48, 4);
	// Go复刻：
	driver := base64Captcha.NewDriverDigit(48, 130, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)

	// 生成 ID, B64s, Answer
	id, b64s, answer, err := captcha.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error("验证码生成失败"))
		return
	}

	// [IMPORTANT] Java 是前端传 key，后端存 Redis。
	// 为了适配，我们读取前端传来的 key。如果前端没传，我们用生成的 id 作为 key 返回给前端。
	key := c.Query("key")
	if key == "" {
		key = id
	}

	// [NEW] 存入 Redis，有效期 5 分钟 (key = "captcha:" + key)
	// Java: redisTemplate.opsForValue().set("captcha:" + key, verCode, 5, TimeUnit.MINUTES);
	err = config.RDB.Set(config.Ctx, "captcha:"+key, strings.ToLower(answer), 5*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Error("Redis 存储失败"))
		return
	}

	// 适配 Java 前端的 SpecCaptcha 输出流，这里返回 JSON 包含图片
	// 如果前端是 <img src="/api/user/captcha?key=xxx">，则应该直接 write b64s 的 decode bytes
	// 假设你的前端能处理 JSON 或者你需要返回图片流：
	// 这里为了通用性，保留 JSON，前端可能需要微调，或者你使用 c.DataFromReader 直接返回图片流
	// 鉴于 Java 代码是 `specCaptcha.out(response.getOutputStream())`，它是直接返回图片的。
	// 这里我们做一个判断，如果 Accept 包含 image，返回图片，否则 JSON (或者直接返回图片)

	// 既然你的 Java 代码是 specCaptcha.out，那前端肯定是把它当图片请求。
	// 我们需要解析 b64s (它包含 data:image/png;base64,前缀)
	// 但 base64Captcha 库稍微有点麻烦。
	// 简单起见，如果前端能改，最好用 JSON。如果不能改，下面是返回图片的逻辑 (略复杂，暂时返回JSON)
	c.JSON(http.StatusOK, gin.H{
		"img": b64s, // 前端可以直接赋值给 <img src>
		"key": key,
	})
}

// [NEW] 发送邮件验证码 (/api/user/sendEmailCode)
func (ctrl *UserController) SendEmailCode(c *gin.Context) {
	// 1. 获取参数
	// Java: @RequestParam String email, captcha, captchaKey, type, username
	email := c.PostForm("email")
	if email == "" {
		email = c.Query("email")
	} // 兼容 Query

	captcha := c.PostForm("captcha")
	if captcha == "" {
		captcha = c.Query("captcha")
	}

	captchaKey := c.PostForm("captchaKey")
	if captchaKey == "" {
		captchaKey = c.Query("captchaKey")
	}

	bizType := c.PostForm("type")
	if bizType == "" {
		bizType = c.Query("type")
	}
	if bizType == "" {
		bizType = "register"
	} // 默认值

	username := c.PostForm("username")
	if username == "" {
		username = c.Query("username")
	}

	// 校验必填
	if email == "" {
		c.JSON(http.StatusOK, utils.Error("邮箱不能为空"))
		return
	}
	if captcha == "" || captchaKey == "" {
		c.JSON(http.StatusOK, utils.Error("请输入图形验证码"))
		return
	}

	// 2. 调用 Service
	err := ctrl.userService.SendEmailCode(email, strings.ToLower(captcha), captchaKey, bizType, username)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Ok().Put("msg", "验证码发送成功"))
}

// [NEW] 注册接口 (/api/user/register)
func (ctrl *UserController) Register(c *gin.Context) {
	// 接收 JSON 参数
	var dto struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Code     string `json:"code"` // 邮件验证码
	}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数格式错误"))
		return
	}

	// 构造 User 对象
	user := &model.User{
		Username: dto.Username,
		Password: dto.Password,
		Email:    dto.Email,
	}

	// 调用 Service
	msg, err := ctrl.userService.Register(user, dto.Code)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Ok().Put("msg", msg))
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
	res.Put("user", user)   // 放入 User 对象
	res.Put("token", token) // 额外给一个 Token (虽然 Java 前端可能只用 user)

	// ⚠️ 关键：Java 前端可能依赖 authorites 字段
	// 你可以在 user model 里加个 Authorities []string 字段并填充

	c.JSON(http.StatusOK, res)
}

// [NEW] 重置密码接口 (/api/user/resetPassword)
func (ctrl *UserController) ResetPassword(c *gin.Context) {
	// 定义 DTO 接收 JSON 参数
	var dto struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"` // 前端传来的新密码
		Code     string `json:"code"`     // 验证码
	}

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusOK, utils.Error("参数格式错误"))
		return
	}

	msg, err := ctrl.userService.ResetPassword(dto.Username, dto.Email, dto.Password, dto.Code)
	if err != nil {
		c.JSON(http.StatusOK, utils.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Ok().Put("msg", msg))
}
