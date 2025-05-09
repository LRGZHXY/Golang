package app

import (
	"common/config"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动程序 启动grpc服务 启动http服务 启用日志 启用数据库
func Run(ctx context.Context) error {
	// 启动grpc服务端
	server := grpc.NewServer()
	go func() {
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			log.Fatalf("user grpc server listen err:%v", err)
		}
		//阻塞操作 server.Serve() 持续监听并处理连接的死循环函数
		err = server.Serve(lis)
		if err != nil {
			log.Fatalf("user grpc server run failed err:%v", err)
		}
	}()
	// 定义匿名函数 赋值给变量stop
	stop := func() {
		server.Stop()
		time.Sleep(3 * time.Second) //暂停三秒
		fmt.Println("stop and finish")
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
				log.Println("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				log.Println("hang up!! user app quit")
				return nil
			default:
				return nil
			}
		}
	}
}
