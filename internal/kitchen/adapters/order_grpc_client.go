package adapters

import (
	"context"
	"github.com/baobao233/gorder/common/genproto/orderpb"
)

type OrderGRPC struct {
	client orderpb.OrderServiceClient
}

func NewOrderGRPC(client orderpb.OrderServiceClient) *OrderGRPC {
	return &OrderGRPC{client: client}
}

func (g *OrderGRPC) UpdateOrder(ctx context.Context, request *orderpb.Order) error {
	_, err := g.client.UpdateOrder(ctx, request)
	return err
}
