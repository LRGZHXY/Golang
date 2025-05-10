package discovery

import "fmt"

// 存储服务的基本信息
type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Version string `json:"version"`
	Weight  int    `json:"weight"`
	Ttl     int64  `json:"ttl"`
}

// BuildRegisterKey 根据服务的Version字段生成一个用于注册服务的键
func (s Server) BuildRegisterKey() string {
	if len(s.Version) == 0 { //没有版本号
		return fmt.Sprintf("/%s/%s", s.Name, s.Addr) // /name/addr
	}
	return fmt.Sprintf("/%s/%s/%s", s.Name, s.Version, s.Addr) // /name/version/addr
}
