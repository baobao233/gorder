package config

import (
	"strings"

	"github.com/spf13/viper"
)

func NewViperConfig() error {
	viper.SetConfigName("global")                       // 指定绑定配置的名字
	viper.SetConfigType("yaml")                         // 指定绑定配置的格式是 yaml 格式
	viper.AddConfigPath("../common/config")             // 指定的规则是：比如 x.go 文件需要用 config.yaml 文件，那就是路径就是从 x.go文件出发到 config.yaml文件的路径
	viper.AddConfigPath("../config")                    // 为 rabbitmq 指定 config.yaml的路径
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_")) // 把环境变量中的中横线变成下横线，并且大写字母，在环境变量中找值，为的是找到 stripe-key，而不是暴露在仓库中
	_ = viper.BindEnv("stripe-key", "STRIPE_KEY", "endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET")
	viper.AutomaticEnv() // 如果有环境变量会在环境变量中去找
	return viper.ReadInConfig()
}
