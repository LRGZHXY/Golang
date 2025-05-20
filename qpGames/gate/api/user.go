package api

import (
	"common"
	"common/biz"
	"common/config"
	"common/jwts"
	"common/logs"
	"common/rpc"
	"context"
	"framework/msError"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"user/pb"
)

type UserHandler struct {
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Register 用户注册
func (u *UserHandler) Register(ctx *gin.Context) {
	//接收参数
	var req pb.RegisterParams
	err2 := ctx.ShouldBindJSON(&req) //将前端传来的JSON数据绑定到req对象中
	if err2 != nil {
		common.Fail(ctx, biz.RequestDataError)
		return
	}
	//调用远程rpc接口
	response, err := rpc.UserClient.Register(context.TODO(), &req)
	if err != nil {
		common.Fail(ctx, msError.ToError(err))
		return
	}
	uid := response.Uid //获取uid
	if len(uid) == 0 {
		common.Fail(ctx, biz.SqlError)
		return
	}
	logs.Info("uid:%s", uid)
	// 生成jwt （A.B.C） A:定义加密算法 B:存储数据 C:签名
	// 构建jwt的claims（声明）
	claims := jwts.CustomClaims{
		Uid: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), //设置过期时间
		},
	}
	token, err := jwts.GenToken(&claims, config.Conf.Jwt.Secret) //生成token
	if err != nil {
		logs.Error("Register jwt gen token err:%v", err)
		common.Fail(ctx, biz.Fail)
		return
	}
	//构造响应数据
	result := map[string]any{
		"token": token,
		"serverInfo": map[string]any{
			"host": config.Conf.Services["connector"].ClientHost,
			"port": config.Conf.Services["connector"].ClientPort,
		},
	}
	common.Success(ctx, result)
}
