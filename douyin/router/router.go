package router

import (
	"douyin/controller"
	"douyin/extra/userinfo"
	"douyin/logger"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Setup(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// 推荐：明确设置可信代理
	r.SetTrustedProxies([]string{"127.0.0.1"}) // 仅信任本机代理

	v1 := r.Group("/api/v1")

	//发送验证码
	v1.GET("/sendCode", controller.SendCode)
	//注册
	v1.POST("/signup", controller.SignUpHandler)
	//登录
	v1.POST("/login", controller.LoginHandler)

	//返回视频id
	v1.GET("/getVideo/:page", controller.GetVideoHandler)
	//返回用户喜欢数据
	v1.GET("/fetch_user_like_videos", userinfo.FetchUserLikeVideos)
	//返回用户主页数据
	v1.GET("/fetch_user_page_videos", userinfo.FetchUserPageVideos)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
