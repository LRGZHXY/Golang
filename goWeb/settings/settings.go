package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"

	"github.com/spf13/viper"
)

func Init() (err error) {
	viper.SetConfigName("config") // 指定配置文件名称（不需要带后缀）
	viper.SetConfigType("yaml")   //指定配置文件类型
	viper.AddConfigPath(".")      //指定查找配置文件的路径（相对路径）
	err = viper.ReadInConfig()    // 读取配置信息
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed,err:%v\n", err)
		return
	}
	viper.WatchConfig()                            //监视配置文件变化
	viper.OnConfigChange(func(in fsnotify.Event) { //配置文件发生变化时会调用此函数
		fmt.Printf("配置文件被修改了")
	})
	return
}
