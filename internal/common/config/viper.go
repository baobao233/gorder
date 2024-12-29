package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

func init() {
	if err := NewViperConfig(); err != nil {
		panic(err)
	}
}

var once sync.Once

// NewViperConfig 抽取成单例模式
func NewViperConfig() (err error) {
	once.Do(func() {
		err = newViperConfig()
	})
	return
}

func newViperConfig() error {
	relPath, err := getRelativePathFromCaller()
	if err != nil {
		return err
	}
	viper.SetConfigName("global") // 指定绑定配置的名字
	viper.SetConfigType("yaml")   // 指定绑定配置的格式是 yaml 格式
	viper.AddConfigPath(relPath)
	viper.EnvKeyReplacer(strings.NewReplacer("-", "_")) // 把环境变量中的中横线变成下横线，并且大写字母，在环境变量中找值，为的是找到 stripe-key，而不是暴露在仓库中
	_ = viper.BindEnv("stripe-key", "STRIPE_KEY", "endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET")
	viper.AutomaticEnv() // 如果有环境变量会在环境变量中去找
	return viper.ReadInConfig()
}

// getRelativePathFromCaller 用于动态添加AddConfigPath中的相对路径，否则需要一个一个调用 AddConfigPath 去添加文件相对路径
func getRelativePathFromCaller() (relPath string, err error) {
	callerPwd, err := os.Getwd()
	if err != nil {
		return
	}
	_, here, _, _ := runtime.Caller(0)
	relPath, err = filepath.Rel(callerPwd, filepath.Dir(here))
	fmt.Printf("caller from %s, here: %s, relpath: %s", callerPwd, here, relPath)
	return
}
