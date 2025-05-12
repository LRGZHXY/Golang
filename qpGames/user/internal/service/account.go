package service

import (
	"common/logs"
	"context"
	"core/repo"
	"user/pb"
)

//创建账号

type AccountService struct {
	pb.UnimplementedUserServiceServer
}

func NewAccountService(manager *repo.Manager) *AccountService {
	return &AccountService{}
}

func (a *AccountService) Register(ctx context.Context, req *pb.RegisterParams) (*pb.RegisterResponse, error) {
	//写注册的业务逻辑

	logs.Info("register server called....")
	return &pb.RegisterResponse{Uid: "10000"}, nil
}
