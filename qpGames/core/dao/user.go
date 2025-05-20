package dao

import (
	"context"
	"core/models/entity"
	"core/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserDao struct {
	repo *repo.Manager
}

func (d *UserDao) FindUserByUid(ctx context.Context, uid string) (*entity.User, error) {
	db := d.repo.Mongo.Db.Collection("user") //获取MongoDB数据库中的user集合
	singleResult := db.FindOne(ctx, bson.D{ //按照uid查找单个用户文档
		{"uid", uid},
	})
	user := new(entity.User)
	err := singleResult.Decode(user) //将查询结果解码到user
	if err != nil {
		if err == mongo.ErrNoDocuments { //用户不存在
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (d *UserDao) Insert(ctx context.Context, user *entity.User) error {
	db := d.repo.Mongo.Db.Collection("user")
	_, err := db.InsertOne(ctx, user) //将传入的user对象插入MongoDB
	return err
}

// UpdateUserAddressByUid 更新用户地址
func (d *UserDao) UpdateUserAddressByUid(ctx context.Context, user *entity.User) error {
	db := d.repo.Mongo.Db.Collection("user")
	_, err := db.UpdateOne(ctx, bson.M{ //按照uid查找用户文档
		"uid": user.Uid,
	}, bson.M{ //更新用户地址
		"$set": bson.M{
			"address":  user.Address,
			"location": user.Location,
		},
	})
	return err
}

func NewUserDao(m *repo.Manager) *UserDao {
	return &UserDao{
		repo: m,
	}
}
