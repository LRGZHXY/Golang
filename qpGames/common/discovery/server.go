package discovery

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// 存储服务的基本信息
type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Weight  int    `json:"weight"`
	Version string `json:"version"`
	Ttl     int64  `json:"ttl"`
}

// BuildRegisterKey 根据服务的Version字段生成一个用于注册服务的键
func (s Server) BuildRegisterKey() string {

	if len(s.Version) == 0 { //没有版本号
		// user
		return fmt.Sprintf("/%s/%s", s.Name, s.Addr) // /name/addr
	}
	return fmt.Sprintf("/%s/%s/%s", s.Name, s.Version, s.Addr) // /name/version/addr
}

// ParseValue 将JSON字符串反序列化为Server类型的对象
func ParseValue(v []byte) (Server, error) {
	var server Server
	if err := json.Unmarshal(v, &server); err != nil {
		return server, err
	}
	return server, nil
}

// ParseKey 从给定的Etcd键中提取name version addr
func ParseKey(key string) (Server, error) {
	strs := strings.Split(key, "/")
	if len(strs) == 2 {
		return Server{
			Name: strs[0],
			Addr: strs[1],
		}, nil
	}
	if len(strs) == 3 {
		return Server{
			Name:    strs[0],
			Addr:    strs[2],
			Version: strs[1],
		}, nil
	}
	return Server{}, errors.New("invalid key")
}
