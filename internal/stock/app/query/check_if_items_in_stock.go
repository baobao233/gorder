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

// TODO: 删掉
var stub = map[string]string{
	"1": "price_1QWx7tGPMmgbVtkIivuPtjEt",
	"2": "price_1QVTiGGPMmgbVtkIR6Aiv7up",
}

func (c checkIfItemsInStockQuery) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	var res []*orderpb.Item
	for _, item := range query.Items {
		// TODO: 改成从数据库获取 or 从stripe 获取
		priceID, ok := stub[item.ID]
		if !ok {
			priceID = stub["1"]
		}
		res = append(res, &orderpb.Item{
			ID:       item.ID,
			Quantity: item.Quantity,
			PriceID:  priceID,
		})
	}
	return res, nil
}
