package connector

import (
	"common/logs"
	"fmt"
	"framework/game"
	"framework/net"
)

type Connector struct {
	isRunning bool
	wsManager *net.Manager
}

func Default() *Connector {
	return &Connector{}
}

func (c *Connector) Run(serverId string) {
	if !c.isRunning {
		//启动websocket和nats
		c.wsManager = net.NewManager()
		c.Server(serverId)
	}
}

func (c *Connector) Close() () {
	if c.isRunning {
		//关闭websocket和nats
		c.wsManager.Close()
	}
}

func (c *Connector) Server(serverId string) {
	logs.Info("run connector:%v", serverId)
	//地址 需要读取配置文件 在游戏中可能加载很多的配置信息 如果写到yml会比较复杂 不容易维护
	//游戏中的配置读取一般采用json的方式 需要读取json的配置文件
	c.wsManager.ServerId = serverId
	connectorConfig := game.Conf.GetConnector(serverId)
	if connectorConfig == nil {
		logs.Fatal("no connector config found")
	}
	addr := fmt.Sprintf("%s:%d", connectorConfig.Host, connectorConfig.ClientPort)
	c.isRunning = true
	c.wsManager.Run(addr)
}
