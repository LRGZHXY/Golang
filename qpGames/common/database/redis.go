package database

import (
	"common/config"
	"common/logs"
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisManager struct {
	Cli        *redis.Client        //存放Redis单机客户端实例
	ClusterCli *redis.ClusterClient //存放Redis集群客户端实例
}

// NewRedis 根据配置初始化Redis客户端
func NewRedis() *RedisManager {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //避免连接、Ping操作无期限等待
	defer cancel()

	var clusterCli *redis.ClusterClient
	var cli *redis.Client
	addrs := config.Conf.Database.RedisConf.ClusterAddrs //从配置文件中读取Redis集群地址
	if len(addrs) == 0 {                                 //创建单机Redis客户端
		cli = redis.NewClient(&redis.Options{
			Addr:         config.Conf.Database.RedisConf.Addr,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns, //最小空闲连接数
			Password:     config.Conf.Database.RedisConf.Password,
		})
	} else { //创建Redis集群客户端
		clusterCli = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Conf.Database.RedisConf.ClusterAddrs,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
		})
	}
	if clusterCli != nil {
		if err := clusterCli.Ping(ctx).Err(); err != nil { //测试连接
			logs.Fatal("redis cluster connect err:%v", err)
			return nil
		}
	}
	if cli != nil {
		if err := cli.Ping(ctx).Err(); err != nil {
			logs.Fatal("redis connect err:%v", err)
			return nil
		}
	}
	return &RedisManager{
		Cli:        cli,
		ClusterCli: clusterCli,
	}
}

// Close 关闭Redis连接
func (r *RedisManager) Close() {
	if r.ClusterCli != nil {
		if err := r.ClusterCli.Close(); err != nil {
			logs.Error("redis cluster close err:%v", err)
		}
	}
	if r.Cli != nil {
		if err := r.Cli.Close(); err != nil {
			logs.Error("redis close err:%v", err)
		}
	}
}

// Set 设置Redis键值对
func (r *RedisManager) Set(ctx context.Context, key, value string, expire time.Duration) error {
	if r.ClusterCli != nil {
		return r.ClusterCli.Set(ctx, key, value, expire).Err()
	}
	if r.Cli != nil {
		return r.Cli.Set(ctx, key, value, expire).Err()
	}
	return nil
}
