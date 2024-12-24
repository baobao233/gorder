package query

import (
	"context"

	"github.com/baobao233/gorder/common/decorator"
	domain "github.com/baobao233/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

// GetCustomerOrder 关注的信息是什么，得到 customer 的 order 的话，那我们就只需要他的 customerID 和 orderID
type GetCustomerOrder struct {
	CustomerID string
	OrderID    string
}

// GetCustomerOrderHandler 起一个别名，实际上是 QueryHandler，用于NewGetCustomerOrderHandler时候返回一个GetCustomerOrderHandler
type GetCustomerOrderHandler decorator.QueryHandler[GetCustomerOrder, *domain.Order]

// 真正实行查询操作逻辑，用于查询的结构体
type getCustomerOrderHandler struct {
	orderRepo domain.Repository
}

func NewGetCustomerOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricClient decorator.MetricClient,
) GetCustomerOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	return decorator.ApplyQueryDecorators[GetCustomerOrder, *domain.Order](
		getCustomerOrderHandler{orderRepo: orderRepo},
		logger,
		metricClient,
	)
}

// Handle 执行查询逻辑
func (g getCustomerOrderHandler) Handle(ctx context.Context, query GetCustomerOrder) (*domain.Order, error) {
	o, err := g.orderRepo.Get(ctx, query.OrderID, query.CustomerID)
	if err != nil {
		return nil, err
	}
	return o, nil
}
