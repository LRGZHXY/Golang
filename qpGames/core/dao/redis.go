package dao

import (
	"context"
	"core/repo"
	"fmt"
)

const Prefix = "MSQP" //Redis键前缀 MSQP:AccountId
const AccountIdRedisKey = "AccountId"
const AccountIdBegin = 10000

type RedisDao struct {
	repo *repo.Manager
}

func (d *RedisDao) NextAccountId() (string, error) {
	//自增 给一个前缀
	return d.incr(Prefix + ":" + AccountIdRedisKey)
}

// incr 自增redis key
func (d *RedisDao) incr(key string) (string, error) {
	todo := context.TODO()
	var exist int64
	var err error
	if d.repo.Redis.Cli != nil { //单节点Redis客户端
		exist, err = d.repo.Redis.Cli.Exists(todo, key).Result() //判断此key是否存在,存在返回1，不存在返回0
	} else {
		exist, err = d.repo.Redis.ClusterCli.Exists(todo, key).Result()
	}
	if exist == 0 {
		//不存在
		if d.repo.Redis.Cli != nil {
			err = d.repo.Redis.Cli.Set(todo, key, AccountIdBegin, 0).Err()
		} else {
			err = d.repo.Redis.ClusterCli.Set(todo, key, AccountIdBegin, 0).Err()
		}
		if err != nil {
			return "", err
		}
	}
	var id int64
	if d.repo.Redis.Cli != nil {
		id, err = d.repo.Redis.Cli.Incr(todo, key).Result() //对key进行自增
	} else {
		id, err = d.repo.Redis.ClusterCli.Incr(todo, key).Result()
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

func NewRedisDao(m *repo.Manager) *RedisDao {
	return &RedisDao{
		repo: m,
	}
}
