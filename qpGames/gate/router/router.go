package router

import (
	"common/config"
	"common/rpc"
	"gate/api"
	"gate/auth"
	"github.com/gin-gonic/gin"
)

// RegisterRouter 注册路由
func RegisterRouter() *gin.Engine {
	if config.Conf.Log.Level == "DEBUG" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	//初始化grpc客户端
	rpc.Init()

	r := gin.Default()
	r.Use(auth.Cors()) //跨域问题
	userHandler := api.NewUserHandler()
	r.POST("/register", userHandler.Register)
	return r
}
