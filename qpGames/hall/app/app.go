package app

import (
	"common/config"
	"common/logs"
	"context"
	"core/repo"
	"framework/node"
	"hall/route"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动程序 启动grpc服务 启动http服务 启用日志 启用数据库
func Run(ctx context.Context, serverId string) error {
	//1.做一个日志库 info error fatal debug..
	logs.InitLog(config.Conf.AppName)
	exit := func() {}
	go func() {
		n := node.Default()
		exit = n.Close
		manager := repo.New()
		n.RegisterHandler(route.Register(manager))
		n.Run(serverId)
	}()
	// 定义匿名函数 赋值给变量stop
	stop := func() {
		exit()
		time.Sleep(3 * time.Second) //暂停三秒
		logs.Info("stop and finish")
	}
	// 优雅启停
	c := make(chan os.Signal, 1) //创建一个channel,用来接收操作系统信号，缓冲区大小为1
	//SIGINT	Ctrl+C 中断程序
	//SIGTERM	kill 命令默认信号
	//SIGQUIT	quit 信号
	//SIGHUP	终端挂起（重启提示）
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	for {
		select { //同时监听多个channel
		case <-ctx.Done(): //上下文被取消
			stop()
			return nil
		case s := <-c: // <- 是接收操作符，表示从 channel 中接收数据
			switch s {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				stop()
				logs.Info("connector app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("hang up!! connector app quit")
				return nil
			default:
				return nil
			}
		}
	}
}
