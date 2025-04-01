package controller

import (
	"douyin/logic"
	"douyin/pkg/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SendCode(c *gin.Context) {
	var p struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	code := util.GenerateVerificationCode()

	util.CacheVerificationCode(p.Email, code)

	if err := util.SendVerificationCode(p.Email); err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

func SignUpHandler(c *gin.Context) {
	var p struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 验证码校验
	cachedCode, exists := util.GetVerificationCode(p.Email)
	if !exists || cachedCode != p.Code {
		ResponseError(c, CodeInvalidParam)
		return
	}

	if err := logic.RegisterUser(p.Nickname, p.Password, p.Email, p.Code); err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

func LoginHandler(c *gin.Context) {
	var p struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	user, err := logic.LoginUser(p.Email, p.Password)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "登录成功", "user": user})
}
