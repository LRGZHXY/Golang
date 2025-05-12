package dao

import (
	"context"
	"core/models/entity"
	"core/repo"
)

type AccountDao struct {
	repo *repo.Manager
}

// SaveAccount 保存账号
func (d *AccountDao) SaveAccount(todo context.Context, ac *entity.Account) error {
	table := d.repo.Mongo.Db.Collection("account") //获取MongoDB数据库中的account集合
	_, err := table.InsertOne(todo, ac)            //保存传入的Account对象
	if err != nil {
		return err
	}
	return nil
}

func NewAccountDao(m *repo.Manager) *AccountDao {
	return &AccountDao{
		repo: m,
	}
}
