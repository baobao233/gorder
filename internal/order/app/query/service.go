package query

import (
	"context"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/common/genproto/stockpb"
)

/*
跟 stock 的通信协议，也就是怎么去调用 stock
*/

// StockService 需要的方法参考stock.proto文件
type StockService interface {
	CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.CheckIfItemsInStockResponse, error)
	GetItems(ctx context.Context, itemIDs []string) ([]*orderpb.Item, error)
}
