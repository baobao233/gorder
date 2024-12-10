package service

import (
	"context"
	"github.com/baobao233/gorder/common/metrics"
	"github.com/baobao233/gorder/order/adapters"
	"github.com/baobao233/gorder/order/app"
	"github.com/baobao233/gorder/order/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(c context.Context) app.Application {
	orderRepo := adapters.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	// 在 CQRS 中肯定是需要用到存储的，所以我们需要把 orderRepo 注入到这里面去，比如有一个 Queries 我们就可以 New 一个东西然后把 orderRepo 传进去实现依赖倒置
	return app.Application{
		Commands: app.Commands{},
		// 注入了一个支持查询的 handler
		Queries: app.Queries{
			GetCustomOrder: query.NewGetCustomerOrderHandler(orderRepo, logger, metricsClient),
		},
	}
}
