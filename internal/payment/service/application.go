package service

import (
	"context"

	grpcClient "github.com/baobao233/gorder/common/client"
	"github.com/baobao233/gorder/common/metrics"
	"github.com/baobao233/gorder/payment/adaptaters"
	"github.com/baobao233/gorder/payment/app"
	"github.com/baobao233/gorder/payment/app/command"
	"github.com/baobao233/gorder/payment/domain"
	"github.com/baobao233/gorder/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	orderClient, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	orderGRPC := adaptaters.NewOrderGRPC(orderClient)
	// memoryProcessor := processor.NewInMemProcessor()
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("stripe-key"))
	return newApplication(ctx, orderGRPC, stripeProcessor), func() {
		_ = closeOrderClient()
	}
}

// 都依赖于接口，因此参数应该是接口
func newApplication(_ context.Context, grpc command.OrderService, processor domain.Processor) app.Application {
	metricsClient := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsConfig{
		Host:        viper.GetString("payment.metrics_export_addr"),
		ServiceName: viper.GetString("payment.service-name"),
	})
	return app.Application{
		Command: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(processor, grpc, logrus.StandardLogger(), metricsClient),
		},
	}
}
