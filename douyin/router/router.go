package router

import (
	"douyin/controller"
	"douyin/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// 推荐：明确设置可信代理
	r.SetTrustedProxies([]string{"127.0.0.1"}) // 仅信任本机代理

	v1 := r.Group("/api/v1")

	//返回视频id
	v1.GET("/getVideo/:page", controller.GetVideoHandler)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
