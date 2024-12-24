package main

import (
	"context"
	"github.com/baobao233/gorder/common/tracing"

	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/discovery"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/common/server"
	"github.com/baobao233/gorder/order/infrastructure/consumer"
	"github.com/baobao233/gorder/order/ports"
	"github.com/baobao233/gorder/order/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("order.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaeger(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	application, cleanup := service.NewApplication(ctx)
	defer cleanup() // 主函数退出时才把连接关闭
	// 初始化消息队列
	ch, closeConn := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = closeConn()
		_ = ch.Close()
	}()

	go consumer.NewConsumer(application).Listen(ch)

	// 注册到 consul 中
	deregisterFunc, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		_ = deregisterFunc()
	}()

	// 启动协程防止阻塞
	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer(application) // 注入 app，类似于胶水层将 handler 和数据库之类的粘合
		// 注册 grpc 服务
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		router.StaticFile("/success", "../../public/success.html")
		// 传入的第二个参数是需要我们自己写的，也就是服务接口的具体实现是什么
		ports.RegisterHandlersWithOptions(router, HTTPServer{
			app: application, // 不要忘记 HTTP 也是需要 app 注入的
		}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
