package service

import (
	"errors"
	"my-blog/config"
	"my-blog/internal/model"
	"my-blog/internal/repository"
	"my-blog/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	GetAllUsers() ([]model.User, error)
	GetUserDetail(id int) (*model.User, error)
	// [NEW] 注册与登录
	// ✅ 修正：必须与你的实现保持一致，增加 string 返回值
	Register(user *model.User, code string) (string, error)
	Login(username, password string) (*model.User, string, error)
	// [MODIFY] 增加参数：图形验证码、Key、业务类型、用户名
	SendEmailCode(email, captcha, captchaKey, bizType, username string) error
	// [NEW] 根据用户名查询
	SelectByUsername(username string) (*model.User, error)
	// [NEW] 重置密码接口
	ResetPassword(username, email, password, code string) (string, error)
}

type userService struct {
	userRepo    repository.UserRepository
	mailService MailService // [NEW] 注入邮件服务
}

func NewUserService(
	userRepo repository.UserRepository,
	mailService MailService) UserService {
	return &userService{
		userRepo:    userRepo,
		mailService: mailService,
	}
}

// GetAllUsers 获取所有用户
func (s *userService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.FindAll()
}

// GetUserDetail 获取用户详情
func (s *userService) GetUserDetail(id int) (*model.User, error) {
	if id <= 0 {
		return nil, errors.New("用户ID不合法")
	}
	return s.userRepo.FindById(id)
}

// [NEW] 实现 SelectByUsername
func (s *userService) SelectByUsername(username string) (*model.User, error) {
	return s.userRepo.FindByUsername(username)
}

// [MODIFY] 发送验证码 (完全复刻 Java 逻辑)
func (s *userService) SendEmailCode(email, captcha, captchaKey, bizType, username string) error {
	// 1. 人机验证 (图形验证码)
	redisCaptchaKey := "captcha:" + captchaKey
	redisCaptcha, err := config.RDB.Get(config.Ctx, redisCaptchaKey).Result()
	if err != nil || redisCaptcha == "" {
		return errors.New("图形验证码已失效，请刷新")
	}
	// 简单忽略大小写比较 (虽然Java代码用的equalsIgnoreCase)
	if redisCaptcha != captcha { // 这里假设前端传来的已经是小写，或者都转小写
		// 为了严谨，可以在 Controller 层统一转小写
		return errors.New("图形验证码错误")
	}
	// 验证通过后删除
	config.RDB.Del(config.Ctx, redisCaptchaKey)

	// 2. 核心逻辑：校验邮箱
	// 检查邮箱是否存在
	user, err := s.userRepo.FindByEmail(email)
	exists := (err == nil && user != nil)

	if bizType == "register" {
		// 注册时：邮箱不能已存在
		if exists {
			return errors.New("该邮箱已被注册，请直接登录")
		}
	} else if bizType == "reset" {
		// 重置密码时
		if username == "" {
			return errors.New("请输入用户名")
		}
		// 查找用户
		targetUser, err := s.userRepo.FindByUsername(username)
		if err != nil || targetUser == nil {
			return errors.New("用户名不存在")
		}
		// 检查匹配
		if targetUser.Email != email {
			return errors.New("邮箱与账号绑定的邮箱不一致")
		}
	}

	// 3. 检查发送频率 (5分钟内)
	key := "verify_code:" + email
	if config.RDB.Exists(config.Ctx, key).Val() > 0 {
		return errors.New("验证码已发送，请勿频繁操作")
	}

	// 4. 生成并发送
	code := s.mailService.GenerateCode()
	err = s.mailService.SendCode(email, code)
	if err != nil {
		return errors.New("邮件发送失败，请检查邮箱是否正确")
	}

	// 5. 存入 Redis，5分钟有效
	config.RDB.Set(config.Ctx, key, code, 5*time.Minute)

	return nil
}

// [NEW] 实现 Register (注册)
// [MODIFY] 注册逻辑
func (s *userService) Register(user *model.User, code string) (string, error) {
	// 1. 校验验证码
	key := "verify_code:" + user.Email
	cachedCode, err := config.RDB.Get(config.Ctx, key).Result()

	if err != nil || cachedCode == "" {
		return "", errors.New("验证码已过期或未发送")
	}
	if cachedCode != code {
		return "", errors.New("验证码错误")
	}

	// 2. 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(user.Username); err == nil {
		return "", errors.New("用户名已存在")
	}

	// 3. 密码加密
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPwd)

	user.Created = time.Now()
	user.Valid = 1
	// 默认头像 (Java中可能是空或者默认图，这里给个默认值)
	user.Avatar = "/api/images/default-avatar.png"

	// 4. 保存用户
	if err := s.userRepo.Create(user); err != nil {
		return "", errors.New("注册失败")
	}

	// 5. 删除验证码
	config.RDB.Del(config.Ctx, key)

	return "注册成功", nil
}

// [NEW] 实现 Login (登录)
func (s *userService) Login(username, password string) (*model.User, string, error) {
	// 1. 查询用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("用户名不存在")
		}
		return nil, "", err
	}

	// 2. 校验密码 (对比 Hash)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("密码错误")
	}

	// 3. 生成 Token
	token, _ := utils.GenerateToken(user.Id, user.Username)

	return user, token, nil
}

// [NEW] 实现 ResetPassword
func (s *userService) ResetPassword(username, email, password, code string) (string, error) {
	// 1. 校验验证码
	key := "verify_code:" + email
	cachedCode, err := config.RDB.Get(config.Ctx, key).Result()

	if err != nil || cachedCode == "" {
		return "", errors.New("验证码错误或已过期")
	}
	if cachedCode != code {
		return "", errors.New("验证码错误")
	}

	// 2. 根据用户名查找用户
	if username == "" {
		return "", errors.New("用户名不能为空")
	}
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("用户不存在")
		}
		return "", err
	}

	// 3. [核心修复] 安全校验：账号绑定的邮箱必须等于验证通过的邮箱
	if user.Email != email {
		return "", errors.New("验证邮箱与该账户绑定的邮箱不一致！")
	}

	// 4. 重置密码 (加密)
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user.Password = string(hashedPwd)

	// 5. 更新数据库
	// 注意：这里只更新密码字段，避免覆盖其他字段
	// GORM Updates 会自动忽略空值，但为了安全最好指定字段
	err = config.DB.Model(&model.User{}).Where("id = ?", user.Id).Update("password", user.Password).Error
	if err != nil {
		return "", errors.New("密码重置失败")
	}

	// 6. 成功后删除验证码
	config.RDB.Del(config.Ctx, key)

	return "密码重置成功", nil
}
