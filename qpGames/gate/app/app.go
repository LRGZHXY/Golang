package app

import (
	"common/config"
	"common/logs"
	"context"
	"fmt"
	"gate/router"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动程序 启动grpc服务 启用http服务  启用日志 启用数据库
func Run(ctx context.Context) error {
	//1.做一个日志库 info error fatal debug
	logs.InitLog(config.Conf.AppName)
	go func() {
		//gin 启动  注册一个路由
		r := router.RegisterRouter()
		//http接口
		if err := r.Run(fmt.Sprintf(":%d", config.Conf.HttpPort)); err != nil {
			logs.Fatal("gate gin run err:%v", err)
		}
	}()
	// 定义匿名函数 赋值给变量stop
	stop := func() {
		time.Sleep(3 * time.Second) //暂停三秒
		logs.Info("stop app finish")
	}
	// 优雅启停
	c := make(chan os.Signal, 1) //创建一个channel,用来接收操作系统信号，缓冲区大小为1
	//SIGINT	Ctrl+C 中断程序
	//SIGTERM	kill 命令默认信号
	//SIGQUIT	quit 信号
	//SIGHUP	终端挂起（重启提示）
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP)
	for {
		select { //同时监听多个channel
		case <-ctx.Done(): //上下文被取消
			stop()
			return nil
		case s := <-c: // <- 是接收操作符，表示从 channel 中接收数据
			switch s {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				stop()
				logs.Info("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("hang up!! user app quit")
				return nil
			default:
				return nil
			}
		}
	}
}
