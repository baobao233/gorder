package service

import (
	"context"
	"github.com/baobao233/gorder/stock/infrastructure/integration"
	"github.com/baobao233/gorder/stock/infrastructure/persistent"
	"github.com/spf13/viper"

	"github.com/baobao233/gorder/common/metrics"
	"github.com/baobao233/gorder/stock/adapters"
	"github.com/baobao233/gorder/stock/app"
	"github.com/baobao233/gorder/stock/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(c context.Context) app.Application {
	//stockRepo := adapters.NewMemoryStockRepository()
	db := persistent.NewMySQL()
	stockRepo := adapters.NewMySQLStockRepository(db)
	stripeAPI := integration.NewStripeAPI()
	metricsClient := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsConfig{
		Host:        viper.GetString("stock.metrics_export_addr"),
		ServiceName: viper.GetString("stock.service-name"),
	})
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logrus.StandardLogger(), metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logrus.StandardLogger(), metricsClient),
		},
	}
}
