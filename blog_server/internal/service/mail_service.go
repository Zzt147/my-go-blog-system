package service

import (
	"fmt"
	"math/rand"
	"my-blog/config"
	"time"

	"gopkg.in/gomail.v2"
)

// [NEW] 邮件服务接口
type MailService interface {
	SendCode(to, code string) error
	GenerateCode() string
}

type mailService struct{}

func NewMailService() MailService {
	return &mailService{}
}

// [NEW] 发送验证码
func (s *mailService) SendCode(to, code string) error {
	m := gomail.NewMessage()
	// 发件人
	m.SetHeader("From", config.Config.Mail.Username)
	// 收件人
	m.SetHeader("To", to)
	// 主题
	m.SetHeader("Subject", "【你的博客名】注册验证码")
	// 正文
	m.SetBody("text/html", fmt.Sprintf("欢迎注册，您的验证码是：<b>%s</b>。有效时间为5分钟，请勿泄露给他人。", code))

	d := gomail.NewDialer(
		config.Config.Mail.Host,
		config.Config.Mail.Port,
		config.Config.Mail.Username,
		config.Config.Mail.Password,
	)

	return d.DialAndSend(m)
}

// [NEW] 生成6位随机数字
func (s *mailService) GenerateCode() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return code
}
