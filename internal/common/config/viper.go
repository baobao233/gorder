package config

import "github.com/spf13/viper"

func NewViperConfig() error {
	viper.SetConfigName("global")           // 指定绑定配置的名字
	viper.SetConfigType("yaml")             // 指定绑定配置的格式是 yaml 格式
	viper.AddConfigPath("../common/config") // 指定后可以在别的文件夹中也去使用
	viper.AutomaticEnv()                    // 如果有环境变量会在环境变量中去找
	return viper.ReadInConfig()
}
