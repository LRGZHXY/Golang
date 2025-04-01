package logic

import (
	"crypto/sha256"
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/pkg/util"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateUserID() string {
	return fmt.Sprintf("%06d", seededRand.Intn(1000000))
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func RegisterUser(nickname, password, email, inputCode string) error {
	cachedCode, exists := util.GetVerificationCode(email)
	if !exists || cachedCode != inputCode {
		return errors.New("验证码错误或已过期")
	}

	_, err := mysql.GetUserByEmail(email)
	if err == nil {
		return errors.New("邮箱已注册")
	}

	user := &models.User{
		ID:       generateUserID(),
		Nickname: nickname,
		Password: hashPassword(password),
		Email:    email,
		//Verified: true,
	}

	return mysql.CreateUser(user)
}

func LoginUser(email, password string) (*models.User, error) {
	user, err := mysql.GetUserByEmail(email)
	if err != nil || user.Password != hashPassword(password) {
		return nil, errors.New("邮箱或密码错误")
	}

	return user, nil
}
