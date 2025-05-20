package common

import (
	"common/biz"
	"framework/msError"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Result struct {
	Code int `json:"code"` //状态码 0成功 非0失败
	Msg  any `json:"msg"`
}

func F(err *msError.Error) Result {
	return Result{
		Code: err.Code,
	}
}
func S(data any) Result {
	return Result{
		Code: biz.OK,
		Msg:  data,
	}
}

// Fail 失败响应函数
func Fail(ctx *gin.Context, err *msError.Error) {
	ctx.JSON(http.StatusOK, Result{
		Code: err.Code,
		Msg:  err.Err.Error(),
	})
}

// Success 成功响应函数
func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Result{
		Code: biz.OK,
		Msg:  data,
	})
}
