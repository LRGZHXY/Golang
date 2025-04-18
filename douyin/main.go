package main

import (
	"douyin/controller"
	"douyin/logger"
	"douyin/router"
	"douyin/settings"
	"fmt"
	"go.uber.org/zap"

	"douyin/dao/mysql"
	"douyin/dao/redis"
)

func main() {
	//1.加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed,err:%v\n", err)
		return
	}
	//2.初始化日志
	if err := logger.Init(); err != nil {
		fmt.Printf("init logger failed,err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success")
	//3.初始化MySQL连接
	if err := mysql.Init(); err != nil {
		fmt.Printf("init mysql failed,err:%v\n", err)
		return
	}
	defer mysql.Close()
	//4.初始化Redis连接
	if err := redis.Init(); err != nil {
		fmt.Printf("init redis failed,err:%v\n", err)
		return
	}
	defer redis.Close()

	// 初始化gin框架内置的校验器使用的翻译器
	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init validator trans failed, err:%v\n", err)
		return
	}
	// 注册路由
	r := router.Setup(settings.Conf.Mode)

	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
	if err != nil {
		fmt.Printf("run server failed, err:%v\n", err)
		return
	}
}
