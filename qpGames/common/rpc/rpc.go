package rpc

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"user/pb"
)

var (
	UserClient pb.UserServiceClient
)

func Init() {
	//etcd解析器 就可以grpc连接的时候 进行触发，通过提供的addr地址 去etcd中进行查找
	r := discovery.NewResolver(config.Conf.Etcd)
	resolver.Register(r)
	userDomain := config.Conf.Domain["user"]
	initClient(userDomain.Name, userDomain.LoadBalance, &UserClient)
}

// initClient 初始化grpc客户端
func initClient(name string, loadBalance bool, client interface{}) {
	//找服务的地址
	addr := fmt.Sprintf("etcd:///%s", name)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials())} //不安全连接
	if loadBalance {                                              //负载均衡
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, "round_robin")))
	}
	conn, err := grpc.DialContext(context.TODO(), addr, opts...) //建立grpc连接
	if err != nil {
		logs.Fatal("rpc connect etcd err:%v", err)
	}
	switch c := client.(type) { //类型断言
	case *pb.UserServiceClient:
		*c = pb.NewUserServiceClient(conn)
	default:
		logs.Fatal("unsupported client type")
	}
}
