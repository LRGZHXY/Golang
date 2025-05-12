package discovery

import (
	"common/config"
	"common/logs"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type Resolver struct {
	conf        config.EtcdConf
	etcdCli     *clientv3.Client //etcd连接
	DialTimeout int              //超时时间
	closeCh     chan struct{}    //关闭信号
	key         string
	cc          resolver.ClientConn
	srvAddrList []resolver.Address
	watchCh     clientv3.WatchChan
}

// Build 构建一个gRPC解析器，通过连接etcd获取服务地址列表并监听其变更，从而实现服务发现与动态负载均衡。
func (r Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc //Resolver通过cc通知gRPC当前有哪些服务节点地址
	//1.连接etcd
	//建立etcd的连接
	var err error
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.conf.Addrs,                               //地址
		DialTimeout: time.Duration(r.DialTimeout) * time.Second, //超时时间
	})
	if err != nil {
		logs.Fatal("grpc client connect etcd err:%v", err)
	}
	r.closeCh = make(chan struct{})
	//2.根据key获取所有服务器地址
	r.key = target.URL.Path
	if err := r.sync(); err != nil {
		return nil, err
	}
	//3.启动监听协程（节点有变动，实时更新信息）
	go r.watch()
	return nil, nil
}

func (r Resolver) Scheme() string {
	return "etcd"
}

// sync 从etcd获取指定服务的地址列表并通知grpc更新连接信息
func (r Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.conf.RWTimeout)*time.Second)
	defer cancel() //避免长时间阻塞

	res, err := r.etcdCli.Get(ctx, r.key, clientv3.WithPrefix()) //从etcd读取所有以r.key为前缀的键值对
	if err != nil {
		logs.Error("grpc client get etcd failed,name=%s,err:%v", r.key, err)
		return err
	}
	r.srvAddrList = []resolver.Address{}
	for _, v := range res.Kvs {
		server, err := ParseValue(v.Value) //遍历服务实例并解析
		if err != nil {
			logs.Error("grpc client parse etcd value failed,name=%s,err:%v", r.key, err)
			continue
		}
		r.srvAddrList = append(r.srvAddrList, resolver.Address{
			Addr:       server.Addr,
			Attributes: attributes.New("weight", server.Weight), //添加权重信息
		})
	}
	if len(r.srvAddrList) == 0 {
		return nil
	}
	err = r.cc.UpdateState(resolver.State{ //更新服务地址列表
		Addresses: r.srvAddrList,
	})
	if err != nil {
		logs.Error("grpc client UpdateState failed,name=%s,err:%v", r.key, err)
		return err
	}
	return nil
}

// watch 监听etcd中的服务地址变化
func (r Resolver) watch() {
	r.watchCh = r.etcdCli.Watch(context.Background(), r.key, clientv3.WithPrefix()) //启动监听
	ticker := time.NewTicker(time.Minute)                                           //启动定时器，1分钟同步一次数据
	for {
		select {
		case <-r.closeCh:
			r.Close() //关闭etcd连接
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				logs.Error("watch sync failed,err:%v", err)
			}
		}
	}
}

// update 根据etcd事件更新服务地址列表
func (r Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		switch ev.Type {
		case clientv3.EventTypePut: //服务注册或变更
			server, err := ParseValue(ev.Kv.Value) //解析新的服务地址
			if err != nil {
				logs.Error("grpc client update(EventTypePut) parse etcd value failed,name=%s,err:%v")
			}
			addr := resolver.Address{
				Addr:       server.Addr,
				Attributes: attributes.New("weight", server.Weight),
			}
			if !Exist(r.srvAddrList, addr) {
				r.srvAddrList = append(r.srvAddrList, addr) //添加到服务地址列表
				err = r.cc.UpdateState(resolver.State{ //通知gRPC框架更新服务地址
					Addresses: r.srvAddrList,
				})
				if err != nil {
					logs.Error("grpc client update(EventTypePut) UpdateState failed,name=%s,err:%v", r.key, err)
				}
			}
		case clientv3.EventTypeDelete: //服务下线
			server, err := ParseKey(string(ev.Kv.Key))
			if err != nil {
				logs.Error("grpc client update(EventTypeDelete) parse etcd value failed,name=%s,err:%v")
			}
			addr := resolver.Address{Addr: server.Addr}
			if list, ok := Remove(r.srvAddrList, addr); ok { //从r.srvAddrList中删除该地址
				r.srvAddrList = list
				err = r.cc.UpdateState(resolver.State{
					Addresses: r.srvAddrList,
				})
				if err != nil {
					logs.Error("grpc client update(EventTypeDelete) UpdateState failed,name=%s,err:%v", r.key, err)
				}
			}
		}
	}
}

// Close 关闭etcd连接
func (r Resolver) Close() {
	if r.etcdCli != nil {
		err := r.etcdCli.Close()
		if err != nil {
			logs.Error("Resolver close etcd err:%v", err)
		}
	}
}

// Exist 判断给定地址是否已经存在于地址列表中
func Exist(list []resolver.Address, addr resolver.Address) bool {
	for i := range list {
		if list[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}

// Remove 从地址列表中删除给定的地址
func Remove(list []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range list {
		if list[i].Addr == addr.Addr {
			list[i] = list[len(list)-1]
			return list[:len(list)-1], true
		}
	}
	return nil, false
}

// NewResolver 创建一个配置为EtcdConf的Resolver实例
func NewResolver(conf config.EtcdConf) *Resolver {
	return &Resolver{
		conf: conf,
	}
}
