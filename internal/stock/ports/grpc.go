package ports

import (
	context "context"
	"github.com/baobao233/gorder/common/tracing"

	"github.com/baobao233/gorder/common/genproto/stockpb"
	"github.com/baobao233/gorder/stock/app"
	"github.com/baobao233/gorder/stock/app/query"
)

/*
用于与外界通信，实现了StockServiceServer接口
*/
type GRPCServer struct {
	app app.Application // 注入 app，app 类似于胶水把数据库，handler之类的粘合起来
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	_, span := tracing.Start(ctx, "GetItems")
	defer span.End()

	items, err := G.app.Queries.GetItems.Handle(ctx, query.GetItems{ItemIDs: request.ItemsIDs})
	if err != nil {
		return nil, err
	}
	return &stockpb.GetItemsResponse{Items: items}, nil
}

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	_, span := tracing.Start(ctx, "CheckIfItemsInStock")
	defer span.End()

	items, err := G.app.Queries.CheckIfItemsInStock.Handle(ctx, query.CheckIfItemsInStock{Items: request.Items})
	if err != nil {
		return nil, err
	}
	return &stockpb.CheckIfItemsInStockResponse{
		Instock: 1,
		Items:   items,
	}, nil
}
