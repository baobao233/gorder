package grpc

import (
	"context"
	"errors"
	"github.com/baobao233/gorder/common/logging"

	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/common/genproto/stockpb"
)

/*
写法比较固定，根据stock_grpc.pb.go生成，作用是实现一个结构体用于给 order 服务调用 stock 的函数
*/

type StockGRPC struct {
	client stockpb.StockServiceClient
}

func NewStockGRPC(client stockpb.StockServiceClient) *StockGRPC {
	return &StockGRPC{client: client}
}

func (s StockGRPC) CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (resp *stockpb.CheckIfItemsInStockResponse, err error) {
	_, dLog := logging.WhenRequest(ctx, "StockGRPC.CheckIfItemsInStock", items)
	defer dLog(resp, &err)

	if items == nil {
		return nil, errors.New("grpc items can not be nil")
	}
	return s.client.CheckIfItemsInStock(ctx, &stockpb.CheckIfItemsInStockRequest{Items: items})
}

func (s StockGRPC) GetItems(ctx context.Context, itemIDs []string) (items []*orderpb.Item, err error) {
	_, dLog := logging.WhenRequest(ctx, "StockGRPC.GetItems", items)
	defer dLog(items, &err)

	resp, err := s.client.GetItems(ctx, &stockpb.GetItemsRequest{ItemsIDs: itemIDs})
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}
