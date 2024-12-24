package service

import (
	"context"

	"github.com/baobao233/gorder/common/metrics"
	"github.com/baobao233/gorder/stock/adapters"
	"github.com/baobao233/gorder/stock/app"
	"github.com/baobao233/gorder/stock/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(c context.Context) app.Application {
	stockRepo := adapters.NewMemoryStockRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, logger, metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
		},
	}
}
