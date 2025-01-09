package main

import (
	"context"
	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/client"
	"github.com/baobao233/gorder/common/tracing"
	"github.com/baobao233/gorder/kitchen/adapters"
	"github.com/baobao233/gorder/kitchen/infrastructure/consumer"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/baobao233/gorder/common/config"
	"github.com/baobao233/gorder/common/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("kitchen.service-name")

	//因为后面需要超时控制，所以需要传入 withcancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaeger(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	orderClient, closeFunc, err := client.NewOrderGRPCClient(ctx) // 通过 consul 发现orderGRPC服务
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		_ = closeFunc()
	}()

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

	orderGRPC := adapters.NewOrderGRPC(orderClient)
	go consumer.NewConsumer(orderGRPC).Listen(ch)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		logrus.Info("received signal, exiting...")
		os.Exit(0)
	}()
	logrus.Infof("to exit, press ctrl+c")
	select {}
}
