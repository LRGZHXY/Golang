package app

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"core/repo"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user/internal/service"
	"user/pb"
)

// Run 启动程序 启动grpc服务 启动http服务 启用日志 启用数据库
func Run(ctx context.Context) error {
	//1.做一个日志库 info error fatal debug..
	logs.InitLog(config.Conf.AppName)

	//2.创建一个新的服务注册对象register
	register := discovery.NewRegister()

	//启动grpc服务端
	server := grpc.NewServer() //创建一个新的gRPC服务器实例server
	//注册grpc service 需要数据库 mongo redis
	//初始化 数据库管理
	manager := repo.New()
	go func() {
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("user grpc server listen err:%v", err)
		}
		//将服务注册到etcd中
		err = register.Register(config.Conf.Etcd)
		if err != nil {
			logs.Fatal("user grpc server register etcd err:%v", err)
		}
		pb.RegisterUserServiceServer(server, service.NewAccountService(manager))
		//阻塞操作 server.Serve() 持续监听并处理连接的死循环函数
		err = server.Serve(lis)
		if err != nil {
			logs.Fatal("user grpc server run failed err:%v", err)
		}
	}()

	// 定义匿名函数 赋值给变量stop
	stop := func() {
		server.Stop()
		register.Close()
		manager.Close()
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
