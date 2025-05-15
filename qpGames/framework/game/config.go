package game

import (
	"common/logs"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

var Conf *Config

const (
	gameConfig = "gameConfig.json"
	servers    = "servers.json"
)

type Config struct {
	GameConfig  map[string]GameConfigValue `json:"gameConfig"`
	ServersConf ServersConf                `json:"serversConf"`
}
type ServersConf struct {
	Nats       NatsConfig         `json:"nats"`
	Connector  []*ConnectorConfig `json:"connector"`
	Servers    []*ServersConfig   `json:"servers"`
	TypeServer map[string][]*ServersConfig
}

type ServersConfig struct {
	ID               string `json:"id"`
	ServerType       string `json:"serverType"`
	HandleTimeOut    int    `json:"handleTimeOut"`
	RPCTimeOut       int    `json:"rpcTimeOut"`
	MaxRunRoutineNum int    `json:"maxRunRoutineNum"`
}

type ConnectorConfig struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	ClientPort int    `json:"clientPort"`
	Frontend   bool   `json:"frontend"`
	ServerType string `json:"serverType"`
}
type NatsConfig struct {
	Url string `json:"url"`
}

type GameConfigValue map[string]any

// InitConfig 读取配置
func InitConfig(configDir string) {
	Conf = new(Config)
	dir, err := os.ReadDir(configDir)
	if err != nil {
		logs.Fatal("read config dir err:%v", err)
	}
	for _, v := range dir {
		configFile := path.Join(configDir, v.Name())
		if v.Name() == gameConfig {
			readGameConfig(configFile)
		}
		if v.Name() == servers {
			readServersConfig(configFile)
		}
	}
}

func readServersConfig(confFile string) {
	var serversConfig ServersConf
	v := viper.New()
	v.SetConfigFile(confFile)
	// 实时监听
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Println("serversConfig配置文件修改了")
		err := v.Unmarshal(&serversConfig)
		if err != nil {
			panic(fmt.Errorf("gameConfig Unmarshal change config data,err:%v \n", err))
		}
		Conf.ServersConf = serversConfig
	})
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("gameConfig读取配置文件出错,err:%v \n", err))
	}
	//解析
	err = v.Unmarshal(&serversConfig)
	if err != nil {
		panic(fmt.Errorf("gameConfig Unmarshal config data,err:%v \n", err))
	}
	Conf.ServersConf = serversConfig
	typeServerConfig()
}

// typeServerConfig 分类存储服务器配置
func typeServerConfig() {
	if len(Conf.ServersConf.Servers) > 0 { //判断是否有服务器配置
		if Conf.ServersConf.TypeServer == nil {
			Conf.ServersConf.TypeServer = make(map[string][]*ServersConfig) //创建map
		}
		for _, v := range Conf.ServersConf.Servers {
			if Conf.ServersConf.TypeServer[v.ServerType] == nil {
				Conf.ServersConf.TypeServer[v.ServerType] = make([]*ServersConfig, 0) //根据服务器类型对配置分组
			}
			Conf.ServersConf.TypeServer[v.ServerType] = append(Conf.ServersConf.TypeServer[v.ServerType], v)
		}
	}
}

// readGameConfig 读取游戏配置
func readGameConfig(confFile string) {
	var gameConfig = make(map[string]GameConfigValue) //存储配置内容的mao
	v := viper.New()
	v.SetConfigFile(confFile)
	// 实时监听
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Println("gameConfig配置文件修改了")
		err := v.Unmarshal(&gameConfig)
		if err != nil {
			panic(fmt.Errorf("gameConfig Unmarshal change config data,err:%v \n", err))
		}
		Conf.GameConfig = gameConfig
	})
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("gameConfig读取配置文件出错,err:%v \n", err))
	}
	//解析
	err = v.Unmarshal(&gameConfig)
	if err != nil {
		panic(fmt.Errorf("gameConfig Unmarshal config data,err:%v \n", err))
	}
	Conf.GameConfig = gameConfig
}

// GetConnector 获取指定id的connector配置
func (c *Config) GetConnector(serverId string) *ConnectorConfig {
	for _, v := range c.ServersConf.Connector {
		if v.ID == serverId {
			return v
		}
	}
	return nil
}
