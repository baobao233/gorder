package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RunHTTPServer(serviceName string, wrapper func(engine *gin.Engine)) {
	// 使用 sub 进入配置的某一层，规范：参数的来源是某个配置
	addr := viper.Sub(serviceName).GetString("http-addr")
	if addr == "" {
		// TODO: Warning log
	}
	RunHTTPServerOnAddr(addr, wrapper)

}

func RunHTTPServerOnAddr(addr string, wrapper func(engine *gin.Engine)) {
	// 主要在里面进行路由的注册和初始化
	// wrapper 的作用在于针对不同服务的 router 进行我们想要传入的 wrapper 进行处理
	apiRouter := gin.New()
	wrapper(apiRouter)
	apiRouter.Group("/api")

	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}
