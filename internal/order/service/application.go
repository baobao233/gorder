package service

import (
	"context"
	"github.com/baobao233/gorder/common/broker"
	grpcClient "github.com/baobao233/gorder/common/client"
	"github.com/baobao233/gorder/common/metrics"
	"github.com/baobao233/gorder/order/adapters"
	"github.com/baobao233/gorder/order/adapters/grpc"
	"github.com/baobao233/gorder/order/app"
	"github.com/baobao233/gorder/order/app/command"
	"github.com/baobao233/gorder/order/app/query"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	stockClient, closeStockClient, err := grpcClient.NewStockGRPCClient(ctx) // 不能用 defer closeStockClient，因为 return 后直接就 conn 关闭了，所以我们需要把closeStockClient返回到主函数中
	if err != nil {
		panic(err)
	}

	// 初始化消息队列
	ch, closeConn := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)

	// 初始化 grpc
	stockGRPC := grpc.NewStockGRPC(stockClient)
	return newApplication(ctx, stockGRPC, ch), func() {
		_ = closeStockClient()
		_ = closeConn()
		_ = ch.Close()
	}
}

func newApplication(_ context.Context, stockGRPC query.StockService, channel *amqp.Channel) app.Application {
	orderRepo := adapters.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	// 在 CQRS 中肯定是需要用到存储的，所以我们需要把 orderRepo 注入到这里面去，比如有一个 Queries 我们就可以 New 一个东西然后把 orderRepo 传进去实现依赖倒置
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderRepo, stockGRPC, channel, logger, metricsClient), // 注入一个支持创建订单的 handler
			UpdateOrder: command.NewUpdateOrderHandler(orderRepo, logger, metricsClient),                     // 注入一个支持更新订单的 handler
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderRepo, logger, metricsClient), // 注入了一个支持查询的 handler
		},
	}
}
