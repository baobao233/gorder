package adaptaters

import (
	"context"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/sirupsen/logrus"
)

type OrderGrpc struct {
	client orderpb.OrderServiceClient
}

func NewOrderGrpc(client orderpb.OrderServiceClient) *OrderGrpc {
	return &OrderGrpc{client: client}
}

func (o OrderGrpc) UpdateOrder(ctx context.Context, order *orderpb.Order) error {
	_, err := o.client.UpdateOrder(ctx, order)
	logrus.Infof("payment_adapter||update_order,err=%v", err)
	return err
}
