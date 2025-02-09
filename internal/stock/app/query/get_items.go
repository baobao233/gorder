package query

import (
	"context"

	"github.com/baobao233/gorder/common/decorator"
	"github.com/baobao233/gorder/common/entity"
	domain "github.com/baobao233/gorder/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type GetItems struct {
	ItemIDs []string
}

type GetItemsHandler decorator.QueryHandler[GetItems, []*entity.Item]

type getItemsHandler struct {
	stockRepo domain.Repository
}

func NewGetItemsHandler(
	stockRepo domain.Repository,
	logger *logrus.Logger,
	metricsClient decorator.MetricClient,
) GetItemsHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	return decorator.ApplyCommandDecorators[GetItems, []*entity.Item](
		getItemsHandler{stockRepo: stockRepo},
		logger,
		metricsClient,
	)
}

func (g getItemsHandler) Handle(ctx context.Context, cmd GetItems) ([]*entity.Item, error) {
	items, err := g.stockRepo.GetItems(ctx, cmd.ItemIDs)
	if err != nil {
		return nil, err
	}
	return items, nil
}
