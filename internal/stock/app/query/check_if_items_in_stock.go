package query

import (
	"context"
	"github.com/baobao233/gorder/common/entity"
	"github.com/baobao233/gorder/common/handler/redis"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/stock/infrastructure/integration"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/baobao233/gorder/common/decorator"
	doamin "github.com/baobao233/gorder/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

const (
	redisLockPrefix = "check_stock_"
)

type CheckIfItemsInStock struct {
	Items []*entity.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*entity.Item]

type checkIfItemsInStockQuery struct {
	stockRepo doamin.Repository      // 用来查询缓存
	stripeAPI *integration.StripeAPI // 用来调取 stripe 获得 priceID
}

func NewCheckIfItemsInStockHandler(
	stockRepo doamin.Repository,
	stripeAPI *integration.StripeAPI,
	logger *logrus.Logger,
	metricsClient decorator.MetricClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	if stripeAPI == nil {
		panic("nil stripeAPI")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*entity.Item](
		checkIfItemsInStockQuery{stockRepo: stockRepo, stripeAPI: stripeAPI},
		logger,
		metricsClient,
	)
}

// Deprecated
var stub = map[string]string{
	"1": "price_1QWx7tGPMmgbVtkIivuPtjEt",
	"2": "price_1QVTiGGPMmgbVtkIR6Aiv7up",
}

func (c checkIfItemsInStockQuery) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*entity.Item, error) {
	if err := lock(ctx, getLockKey(query)); err != nil {
		return nil, errors.Wrapf(err, "redis lock error, key=%s", getLockKey(query))
	}
	defer func() {
		if err := unlock(ctx, getLockKey(query)); err != nil {
			logging.Warnf(ctx, nil, "redis unlock failed, err=%v", err)
		}
	}()
	if err := c.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	var res []*entity.Item
	for _, item := range query.Items {
		priceID, err := c.stripeAPI.GetPriceByProductID(ctx, item.ID) // 需要确保 item.ID和 stripe 中的 ProductID 相对应
		if err != nil || priceID == "" {
			return nil, err
		}
		res = append(res, entity.NewItem(item.ID, "", item.Quantity, priceID))
	}
	return res, nil
}

func getLockKey(query CheckIfItemsInStock) string {
	var ids []string
	for _, i := range query.Items {
		ids = append(ids, i.ID)
	}
	return redisLockPrefix + strings.Join(ids, "_")
}

func lock(ctx context.Context, key string) error {
	return redis.SetNX(ctx, redis.LocalClient(), key, "1", 5*time.Minute)
}

func unlock(ctx context.Context, key string) error {
	return redis.Del(ctx, redis.LocalClient(), key)
}

func (c checkIfItemsInStockQuery) checkStock(ctx context.Context, query []*entity.ItemWithQuantity) error {
	var ids []string
	for _, item := range query {
		ids = append(ids, item.ID)
	}
	records, err := c.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}

	// 记录库存，用于比较
	idQuantityMap := make(map[string]int32)
	for _, r := range records {
		idQuantityMap[r.ID] += r.Quantity
	}

	var (
		ok       = true
		failedOn []struct {
			ID   string
			Want int32
			Have int32
		}
	)
	for _, item := range query {
		if item.Quantity > idQuantityMap[item.ID] {
			ok = false
			failedOn = append(failedOn, struct {
				ID   string
				Want int32
				Have int32
			}{ID: item.ID, Want: item.Quantity, Have: idQuantityMap[item.ID]})
		}
	}

	// 更新库存数量逻辑在 func 中，传入给transaction 中去执行
	if ok {
		return c.stockRepo.UpdateStock(ctx, query, func(
			ctx context.Context,
			existing []*entity.ItemWithQuantity,
			query []*entity.ItemWithQuantity,
		) ([]*entity.ItemWithQuantity, error) {
			var newItems []*entity.ItemWithQuantity
			for _, e := range existing {
				for _, q := range query {
					if e.ID == q.ID {
						iq, err := entity.NewValidItemWithQuantity(e.ID, e.Quantity-q.Quantity)
						if err != nil {
							return nil, err
						}
						newItems = append(newItems, iq)
					}
				}
			}
			return newItems, nil
		})
	}
	return doamin.ExceedStockError{FailedOn: failedOn}
}

func getStubPriceID(id string) string {
	priceID, ok := stub[id]
	if !ok {
		priceID = stub["1"]
	}
	return priceID
}
