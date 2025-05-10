package database

import (
	"common/config"
	"common/logs"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type MongoManager struct {
	Cli *mongo.Client
	Db  *mongo.Database
}

func NewMongo() *MongoManager {
	//创建了一个带有超时时间的上下文(防止MongoDB连接操作时间过长,10秒没连上就自动取消)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//连接mongodb
	//创建一个MongoDB客户端配置对象，并根据配置文件里的MongoDB URL自动解析并设置连接参数
	clientOptions := options.Client().ApplyURI(config.Conf.Database.MongoConf.Url)
	clientOptions.SetAuth(options.Credential{ //设置认证信息
		Username: config.Conf.Database.MongoConf.UserName,
		Password: config.Conf.Database.MongoConf.Password,
	})
	clientOptions.SetMinPoolSize(uint64(config.Conf.Database.MongoConf.MinPoolSize)) //设置连接池大小
	clientOptions.SetMaxPoolSize(uint64(config.Conf.Database.MongoConf.MaxPoolSize))
	client, err := mongo.Connect(ctx, clientOptions) //连接mongodb
	if err != nil {
		logs.Fatal("mongo connect failed,err:%v", err)
		return nil
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil { //测试mongodb是否可用
		logs.Fatal("mongo ping failed,err:%v", err)
		return nil
	}
	m := &MongoManager{
		Cli: client, //赋值Cli字段为client
	}
	m.Db = m.Cli.Database(config.Conf.Database.MongoConf.Db)
	return m
}

func (m *MongoManager) Close() {
	err := m.Cli.Disconnect(context.TODO()) //断开与MongoDB的连接
	if err != nil {
		logs.Error("mongo close err:%v", err)
	}
}
