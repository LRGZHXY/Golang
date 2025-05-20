package remote

import (
	"common/logs"
	"framework/game"
	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	serverId string
	conn     *nats.Conn
	readChan chan []byte
}

func NewNatsClient(serverId string, readChan chan []byte) *NatsClient {
	return &NatsClient{
		serverId: serverId,
		readChan: readChan,
	}
}

func (c *NatsClient) Run() error {
	var err error
	c.conn, err = nats.Connect(game.Conf.ServersConf.Nats.Url) //连接nats服务器
	if err != nil {
		logs.Error("connect nats server fail,err:%v", err)
		return err
	}
	go c.sub() //启动订阅协程
	return nil
}

func (c *NatsClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// sub 接收其他服务通过nats向当前服务发送的消息
func (c *NatsClient) sub() {
	_, err := c.conn.Subscribe(c.serverId, func(msg *nats.Msg) {
		c.readChan <- msg.Data
	})
	if err != nil {
		logs.Error("nats sub err:%v", err)
	}
}

// SendMsg 通过nats向指定的目标服务发送消息
func (c *NatsClient) SendMsg(dst string, data []byte) error {
	if c.conn != nil {
		return c.conn.Publish(dst, data) //将消息发送到主题为dst的订阅者
	}
	return nil
}
