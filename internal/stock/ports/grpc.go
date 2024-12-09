package ports

import (
	context "context"
	"github.com/baobao233/gorder/common/genproto/stockpb"
	"github.com/baobao233/gorder/stock/app"
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
	//TODO implement me
	panic("implement me")
}

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	//TODO implement me
	panic("implement me")
}
