package server

import (
	"github.com/baobao233/gorder/common/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func RunHTTPServer(serviceName string, wrapper func(engine *gin.Engine)) {
	// 使用 sub 进入配置的某一层，规范：参数的来源是某个配置
	addr := viper.Sub(serviceName).GetString("http-addr")
	if addr == "" {
		panic("empty http address")
	}
	RunHTTPServerOnAddr(addr, wrapper)

}

func RunHTTPServerOnAddr(addr string, wrapper func(engine *gin.Engine)) {
	// 主要在里面进行路由的注册和初始化
	// wrapper 的作用在于针对不同服务的 router 进行我们想要传入的 wrapper 进行处理
	apiRouter := gin.New()
	setMiddlewares(apiRouter)
	wrapper(apiRouter)
	apiRouter.Group("/api")

	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}

func setMiddlewares(r *gin.Engine) {
	r.Use(middlewares.StructureLogger(logrus.NewEntry(logrus.StandardLogger())))
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("default_server "))
}
