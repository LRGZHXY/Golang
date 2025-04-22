package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	pb "grpc-study/hello-server/proto"
	"net"
)

type server struct {
	pb.UnimplementedSayHelloServer
}

// 业务
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	//获取元数据信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("未传输token")
	}
	var appId string
	var appKey string
	if v, ok := md["appid"]; ok {
		appId = v[0]
	}
	if v, ok := md["appkey"]; ok {
		appKey = v[0]
	}

	if appId != "kuangshen" || appKey != "123123" {
		return nil, errors.New("token不正确")
	}

	fmt.Printf("hello" + req.RequestName)
	return &pb.HelloResponse{ResponseMsg: "hello" + req.RequestName}, nil
}

func main() {
	//TSL认证
	/*creds, _ := credentials.NewServerTLSFromFile("E:\\GoPractice\\grpc-study\\key\\test.pem",
	"E:\\GoPractice\\grpc-study\\key\\test.key")*/

	//开启端口
	listen, _ := net.Listen("tcp", ":9090")
	//创建grpc服务
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	//grpcServer := grpc.NewServer(grpc.Creds(creds))
	//在grpc服务端注册自己编写的服务
	pb.RegisterSayHelloServer(grpcServer, &server{})
	//启动服务
	err := grpcServer.Serve(listen)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}
