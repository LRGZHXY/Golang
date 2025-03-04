package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func SignUpHandler(c *gin.Context) {
	//1.获取参数和参数校验
	var p models.ParamSignUp
	if err := c.ShouldBindJSON(&p); err != nil {
		//请求参数有误,直接返回响应
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"msg": "请求参数有误",
		})
		return
	}
	fmt.Println(p)
	//2.业务处理
	logic.SignUp()
	//3.返回响应
	c.JSON(http.StatusOK, gin.H{
		"msg": "success",
	})
}
