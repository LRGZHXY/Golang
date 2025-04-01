package util

import (
	"douyin/settings"
	"fmt"
	"gopkg.in/gomail.v2"
	"math/rand"
	"sync"
	"time"
)

var (
	codeCache = make(map[string]struct {
		Code      string
		ExpiresAt time.Time
	})
	mu sync.Mutex
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 生成 6 位验证码
func GenerateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// 缓存验证码（不返回错误）
func CacheVerificationCode(email string, code string) {
	// Example caching logic
	mu.Lock()
	defer mu.Unlock()

	codeCache[email] = struct {
		Code      string
		ExpiresAt time.Time
	}{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute), // Cache expiration time
	}
}

// 获取验证码
func GetVerificationCode(email string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	data, exists := codeCache[email]
	if !exists || time.Now().After(data.ExpiresAt) {
		return "", false
	}
	return data.Code, true
}

// 发送验证码邮件
func SendVerificationCode(email string) error {
	code := GenerateVerificationCode()
	CacheVerificationCode(email, code)

	msg := gomail.NewMessage()
	msg.SetHeader("From", settings.Conf.SMTPConfig.User)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "抖音账号注册验证码")
	msg.SetBody("text/plain", "您的验证码是："+code)

	dialer := gomail.NewDialer(
		settings.Conf.SMTPConfig.Host,
		settings.Conf.SMTPConfig.Port,
		settings.Conf.SMTPConfig.User,
		settings.Conf.SMTPConfig.Password,
	)

	return dialer.DialAndSend(msg)
}
