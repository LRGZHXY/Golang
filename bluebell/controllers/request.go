package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const CtxUserIdKey = "userID"

var ErrorUserNotLogin = errors.New("用户未登录")

// getCurrentUserID 获取当前登录用户的ID
func getCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(CtxUserIdKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	userID, ok = uid.(int64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}
	return
}
