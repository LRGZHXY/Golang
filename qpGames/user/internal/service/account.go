package service

import (
	"common/biz"
	"context"
	"core/dao"
	"core/models/entity"
	"core/models/requests"
	"core/repo"
	"framework/msError"
	"time"
	"user/pb"
)

//创建账号

type AccountService struct {
	accountDao *dao.AccountDao
	redisDao   *dao.RedisDao
	pb.UnimplementedUserServiceServer
}

func NewAccountService(manager *repo.Manager) *AccountService {
	return &AccountService{
		accountDao: dao.NewAccountDao(manager),
		redisDao:   dao.NewRedisDao(manager),
	}
}

// Register 用户注册
func (a *AccountService) Register(ctx context.Context, req *pb.RegisterParams) (*pb.RegisterResponse, error) {
	//写注册的业务逻辑
	if req.LoginPlatform == requests.WeiXin { //是否是微信注册
		ac, err := a.wxRegister(req)
		if err != nil {
			return &pb.RegisterResponse{}, msError.GrpcError(err)
		}
		return &pb.RegisterResponse{
			Uid: ac.Uid, //注册成功返回uid
		}, nil
	}
	return &pb.RegisterResponse{}, nil
}

// wxRegister 通过微信注册
func (a *AccountService) wxRegister(req *pb.RegisterParams) (*entity.Account, *msError.Error) {
	ac := &entity.Account{ //构造Account对象
		WxAccount:  req.Account,
		CreateTime: time.Now(),
	}
	uid, err := a.redisDao.NextAccountId() //生成uid
	if err != nil {
		return ac, biz.SqlError
	}
	ac.Uid = uid
	err = a.accountDao.SaveAccount(context.TODO(), ac) //将账户信息保存到数据库
	if err != nil {
		return ac, biz.SqlError
	}
	return ac, nil
}
