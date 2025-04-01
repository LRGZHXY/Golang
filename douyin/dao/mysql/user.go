package mysql

import (
	"douyin/models"
	"go.uber.org/zap"
)

func CreateUser(user *models.User) error {
	query := "INSERT INTO user (id, nickname, email, password) VALUES (:id, :nickname, :email, :password)"
	_, err := db.NamedExec(query, user)
	if err != nil {
		zap.L().Error("CreateUser failed", zap.Error(err))
	}
	return err
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := "SELECT * FROM user WHERE email = ? LIMIT 1"
	err := db.Get(&user, query, email)
	if err != nil {
		zap.L().Error("GetUserByEmail failed", zap.Error(err))
		return nil, err
	}
	return &user, nil
}
