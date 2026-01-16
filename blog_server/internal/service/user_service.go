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
	Register(user *model.User, code string) error
	Login(username, password string) (*model.User, string, error)
	SendEmailCode(email string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
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

// [NEW] 实现 SendEmailCode (发送验证码)
func (s *userService) SendEmailCode(email string) error {
	// 1. 检查 Redis 是否已有验证码 (防止频繁发送)
	key := "verify_code:" + email
	if config.RDB.Exists(config.Ctx, key).Val() > 0 {
		return errors.New("验证码已发送，请勿频繁操作")
	}

	// 2. 检查邮箱是否已被注册
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return errors.New("该邮箱已被注册")
	}

	// 3. 生成验证码 (这里简单生成 6 位随机数，实际可以用 math/rand)
	code := "123456" // ⚠️ 暂时写死方便测试，你可以写个随机函数替换它
	
	// 4. 发送邮件 (暂时模拟，等引入gomail后替换)
	// mockSendMail(email, code)
	println("📧 [模拟邮件发送] To:", email, "Code:", code)

	// 5. 存入 Redis (5分钟有效)
	config.RDB.Set(config.Ctx, key, code, 5*time.Minute)
	return nil
}

// [NEW] 实现 Register (注册)
func (s *userService) Register(user *model.User, code string) error {
	// 1. 校验验证码
	key := "verify_code:" + user.Email
	redisCode, err := config.RDB.Get(config.Ctx, key).Result()
	if err != nil || redisCode != code {
		return errors.New("验证码错误或已过期")
	}

	// 2. 检查用户名是否存在
	if _, err := s.userRepo.FindByUsername(user.Username); err == nil {
		return errors.New("用户名已存在")
	}

	// 3. 密码加密 (BCrypt)
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPwd)
	
	user.Created = time.Now()
	user.Valid = 1
	// 默认头像
	user.Avatar = "/api/images/default-avatar.png"

	// 4. 保存用户
	if err := s.userRepo.Create(user); err != nil {
		return err
	}
	
	// 5. 删除验证码
	config.RDB.Del(config.Ctx, key)
	return nil
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