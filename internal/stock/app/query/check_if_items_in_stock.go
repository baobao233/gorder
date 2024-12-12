package query

import (
	"context"
	"github.com/baobao233/gorder/common/decorator"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	doamin "github.com/baobao233/gorder/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStockQuery struct {
	stockRepo doamin.Repository
}

func NewCheckIfItemsInStockHandler(
	stockRepo doamin.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStockQuery{stockRepo: stockRepo},
		logger,
		metricsClient,
	)
}

func (c checkIfItemsInStockQuery) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	var res []*orderpb.Item
	for _, item := range query.Items {
		res = append(res, &orderpb.Item{
			ID:       item.ID,
			Quantity: item.Quantity,
		})
	}
	return res, nil
}
