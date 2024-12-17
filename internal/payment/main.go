package main

import (
	"context"
	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/common/server"
	"github.com/baobao233/gorder/payment/infrastructure/consumer"
	"github.com/baobao233/gorder/payment/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serviceName := viper.GetString("payment.service-name")
	serviceType := viper.GetString("payment.server-to-run")

	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

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

	// payment 服务需要启动一个协程监听 channel，也就是消费者; 另外consumer还需要用到application，因此需要把 application 传入到 consumer 中
	go consumer.NewConsumer(application).Listen(ch)

	paymentHandler := NewPaymentHandler()
	switch serviceType {
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported server type: grpc") //用 panic 而不是 fatal 的原因是我们还有一些清理函数需要我们在 defer 中调用
	default:
		logrus.Panic("unsupported server")
	}
}
