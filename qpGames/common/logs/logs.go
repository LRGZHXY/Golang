package logs

import (
	"common/config"
	"github.com/charmbracelet/log"
	"os"
	"time"
)

var logger *log.Logger

// InitLog 初始化日志
func InitLog(appName string) {
	logger = log.New(os.Stderr)
	if config.Conf.Log.Level == "DEBUG" { //控制日志显示的详细程度
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}
	logger.SetPrefix(appName)           //设置日志前缀
	logger.SetReportTimestamp(true)     //启用时间戳，每条日志会显示时间
	logger.SetTimeFormat(time.DateTime) //时间戳格式
}

func Fatal(format string, values ...any) {
	if len(values) == 0 {
		logger.Fatal(format)
	} else {
		logger.Fatalf(format, values) //会按照格式化字符串输出,支持 %v, %s, %d 等格式化占位符
	}
}

func Info(format string, values ...any) {
	if len(values) == 0 {
		logger.Info(format)
	} else {
		logger.Infof(format, values)
	}
}

func Warn(format string, values ...any) {
	if len(values) == 0 {
		logger.Warn(format)
	} else {
		logger.Warnf(format, values)
	}
}

func Debug(format string, values ...any) {
	if len(values) == 0 {
		logger.Debug(format)
	} else {
		logger.Debugf(format, values)
	}
}

func Error(format string, values ...any) {
	if len(values) == 0 {
		logger.Error(format)
	} else {
		logger.Errorf(format, values)
	}
}
