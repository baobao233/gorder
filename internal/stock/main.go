package main

import (
	"github.com/baobao233/gorder/common/genproto/stockpb"
	"github.com/baobao233/gorder/common/server"
	"github.com/baobao233/gorder/stock/ports"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")
	switch serverType {
	case "grpc":
		// 传入 registerFunction是说告诉这个服务注册到哪儿，也是在我们使用 codegen 生成的代码那里面去找
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer()
			stockpb.RegisterStockServiceServer(server, svc) // 传入的第二个参数是需要我们自己写的服务逻辑，也就是服务接口的具体实现是什么
		})
	case "http":
	// 暂时不用
	default:
		panic("unexpected server type")
	}

}
