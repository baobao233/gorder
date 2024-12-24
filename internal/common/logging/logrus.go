package logging

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

/*
该文件用于对日志进行标准化的输出，看起来美观
*/

func Init() {
	SetFormatter(logrus.StandardLogger())
	logrus.SetLevel(logrus.DebugLevel) // 开发环境中设置为 debug 模式
}

func SetFormatter(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyMsg:   "message",
		},
	})
	// 如果是在本地模式中，强制格式化
	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
		//logrus.SetFormatter(&prefixed.TextFormatter{
		//	ForceColors: false,
		//})
	}
}
