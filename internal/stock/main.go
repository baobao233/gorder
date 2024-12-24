package main

import (
	"context"
	"github.com/baobao233/gorder/common/tracing"

	"github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/discovery"
	"github.com/baobao233/gorder/common/genproto/stockpb"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/common/server"
	"github.com/baobao233/gorder/stock/ports"
	"github.com/baobao233/gorder/stock/service"
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
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")

	//因为后面需要超时控制，所以需要传入 withcancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaeger(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	application := service.NewApplication(ctx)

	// 注册到 consul 中
	deregisterFunc, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		_ = deregisterFunc()
	}()

	switch serverType {
	case "grpc":
		// 传入 registerFunction是说告诉这个服务注册到哪儿，也是在我们使用 codegen 生成的代码那里面去找
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer(application)         //完成依赖注入, 令 grpc 中有 application
			stockpb.RegisterStockServiceServer(server, svc) // 传入的第二个参数是需要我们自己写的服务逻辑，也就是服务接口的具体实现是什么
		})
	case "http":
	// 暂时不用
	default:
		panic("unexpected server type")
	}

}
