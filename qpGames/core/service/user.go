package service

import (
	"common/biz"
	"common/logs"
	"common/utils"
	"connector/models/request"
	"context"
	"core/dao"
	"core/models/entity"
	"core/repo"
	"fmt"
	"framework/game"
	"framework/msError"
	hall "hall/models/request"
	"time"
)

type UserService struct {
	userDao *dao.UserDao
}

func (s *UserService) FindAndSaveUserByUid(ctx context.Context, uid string, info request.UserInfo) (*entity.User, error) {
	user, err := s.userDao.FindUserByUid(ctx, uid) //查询mongo中是否存在该用户
	if err != nil {
		logs.Error("[UserService] FindAndSaveUserByUid  user err:%v", err)
		return nil, err
	}
	if user == nil { //如果用户不存在，则创建一个新用户
		user = &entity.User{}
		user.Uid = uid
		user.Gold = int64(game.Conf.GameConfig["startGold"]["value"].(float64))        //初始金币
		user.Avatar = utils.Default(info.Avatar, "Common/head_icon_default")           //头像
		user.Nickname = utils.Default(info.Nickname, fmt.Sprintf("%s%s", "码神", uid)) //昵称
		user.Sex = info.Sex                                                            //性别 0 男 1 女
		user.CreateTime = time.Now().UnixMilli()                                       //创建时间
		user.LastLoginTime = time.Now().UnixMilli()                                    //最后登录时间
		err = s.userDao.Insert(context.TODO(), user)                                   //插入mongo
		if err != nil {
			logs.Error("[UserService] FindAndSaveUserByUid insert user err:%v", err)
			return nil, err
		}
	}
	return user, nil
}

// FindUserByUid 根据uid查询用户信息
func (s *UserService) FindUserByUid(ctx context.Context, uid string) (*entity.User, *msError.Error) {
	user, err := s.userDao.FindUserByUid(ctx, uid)
	if err != nil {
		logs.Error("[UserService] FindUserByUid  user err:%v", err)
		return nil, biz.SqlError
	}
	return user, nil
}

// UpdateUserAddressByUid 更新用户地址
func (s *UserService) UpdateUserAddressByUid(uid string, req hall.UpdateUserAddressReq) error {
	user := &entity.User{ //构建user实体
		Uid:      uid,
		Address:  req.Address,
		Location: req.Location,
	}
	err := s.userDao.UpdateUserAddressByUid(context.TODO(), user)
	if err != nil {
		logs.Error("userDao.UpdateUserAddressByUid err:%v", err)
		return err
	}
	return nil
}

func NewUserService(r *repo.Manager) *UserService {
	return &UserService{
		userDao: dao.NewUserDao(r),
	}
}
