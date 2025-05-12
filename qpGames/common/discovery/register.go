package discovery

import (
	"common/config"
	"common/logs"
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Register grpc服务注册到etcd
// 原理：创建一个租约，将服务信息注册到etcd中，绑定租约
// 过了租约时间，etcd会自动删除服务信息
// 实现心跳，完成续租，如果etcd没有 就新注册
type Register struct {
	etcdCli       *clientv3.Client                        //etcd连接
	leaseId       clientv3.LeaseID                        //租约id
	DialTimeout   int                                     //超时时间
	ttl           int64                                   //租约时间
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse //心跳
	info          Server                                  //注册的服务信息
	closeCh       chan struct{}                           //关闭信号
}

// createLease 创建租约
// ttl 租约时间 单位秒
func (r *Register) createLease(ctx context.Context, ttl int64) error {
	grant, err := r.etcdCli.Grant(ctx, ttl)
	if err != nil {
		logs.Error("createLease failed,err:%v", err)
		return err
	}
	r.leaseId = grant.ID
	return nil
}

// bindLease 绑定租约
func (r *Register) bindLease(ctx context.Context, key, value string) error {
	_, err := r.etcdCli.Put(ctx, key, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		logs.Error("bindLease failed,err:%v", err)
		return err
	}
	logs.Info("register service success,key=%s", key)
	return nil
}

// keepAlive 心跳 续租
func (r *Register) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	//注意是context.Background() 永不过期
	//心跳 要求是一个长连接 如果做了超时 长连接就断掉了 不要设置超时
	//就是一直不停的发消息 保持租约 续租
	keepAliveResponses, err := r.etcdCli.KeepAlive(context.Background(), r.leaseId)
	if err != nil {
		logs.Error("keepAlive failed,err:%v", err)
		return keepAliveResponses, err
	}
	return keepAliveResponses, nil
}

// watcher 监控协程
//1.是否收到关闭信号
//2.租约是否失效
//3.定时检查续约状态 每ttl秒触发一次
func (r *Register) watcher() {
	//定期检查租约是否到期
	ticker := time.NewTicker(time.Duration(r.info.Ttl) * time.Second) //创建一个定时器
	for {
		select {
		case <-r.closeCh: //收到关闭信号
			//注销服务
			if err := r.unregister(); err != nil {
				logs.Error("close and unregister failed,err:%v", err)
			}
			//撤销租约
			if _, err := r.etcdCli.Revoke(context.Background(), r.leaseId); err != nil {
				logs.Error("close and Revoke lease failed,err:%v", err)
			}
			if r.etcdCli != nil {
				r.etcdCli.Close()
			}
			logs.Info("unregister etcd...")
		case <-r.keepAliveChan: //收到续约信号
			//logs.Info("续约成功，%v", res)
			//续约
			/*if res == nil { // "=="!!! 不然会死循环发送“user: register service success,key=/user/v1/127.0.0.1:11500”
				if err := r.register(); err != nil {
					logs.Error("keepAliveChan register failed,err:%v", err)
				}
			}*/
		case <-ticker.C: //定时器触发
			if r.keepAliveChan == nil {
				if err := r.register(); err != nil {
					logs.Error("ticker register failed,err:%v", err)
				}
			}
		}
	}
}

// register 把服务注册到etcd
func (r *Register) register() error {
	//1.创建租约
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	var err error
	if err = r.createLease(ctx, r.info.Ttl); err != nil {
		return err
	}
	//2.开启续租（心跳）
	if r.keepAliveChan, err = r.keepAlive(); err != nil {
		return err
	}
	//3.绑定租约
	data, err := json.Marshal(r.info) //将r.info转成JSON字符串
	if err != nil {
		logs.Error("etcd register json marshal error:%v", err)
		return err
	}
	return r.bindLease(ctx, r.info.BuildRegisterKey(), string(data))
}

// Register 把一个服务注册到etcd
func (r *Register) Register(conf config.EtcdConf) error {
	//组装注册信息
	info := Server{
		Name:    conf.Register.Name,
		Addr:    conf.Register.Addr,   //服务地址
		Weight:  conf.Register.Weight, //权重
		Version: conf.Register.Version,
		Ttl:     conf.Register.Ttl,
	}
	//建立etcd的连接
	var err error
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Addrs,                                 //地址
		DialTimeout: time.Duration(r.DialTimeout) * time.Second, //超时时间
	})
	if err != nil {
		return err
	}
	//注册
	r.info = info
	if err = r.register(); err != nil {
		return err
	}
	//创建一个closeCh通道,用来通知关闭watcher协程
	r.closeCh = make(chan struct{})
	//启动守护协程,保证注册信息的租约一直存在
	go r.watcher()
	return nil
}

// unregister 注销
func (r *Register) unregister() error {
	_, err := r.etcdCli.Delete(context.Background(), r.info.BuildRegisterKey())
	return err
}

// NewRegister 创建一个register实例
func NewRegister() *Register {
	return &Register{
		DialTimeout: 3, //默认超时3秒
	}
}

// Close 关闭Register实例的操作
func (r *Register) Close() {
	r.closeCh <- struct{}{}
}
