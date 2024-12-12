package main

import (
	"context"
	"github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/common/server"
	"github.com/baobao233/gorder/order/ports"
	"github.com/baobao233/gorder/order/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("order.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application, cleanup := service.NewApplication(ctx)
	defer cleanup() // 主函数退出时才把连接关闭

	// 启动协程防止阻塞
	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer(application) // 注入 app，类似于胶水层将 handler 和数据库之类的粘合
		// 注册 grpc 服务
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
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
