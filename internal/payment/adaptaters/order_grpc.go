package adaptaters

import (
	"context"
	"github.com/baobao233/gorder/common/tracing"
	"google.golang.org/grpc/status"

	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/sirupsen/logrus"
)

type OrderGRPC struct {
	client orderpb.OrderServiceClient
}

func NewOrderGRPC(client orderpb.OrderServiceClient) *OrderGRPC {
	return &OrderGRPC{client: client}
}

func (o OrderGRPC) UpdateOrder(ctx context.Context, order *orderpb.Order) (err error) {
	defer func() {
		if err != nil {
			logrus.Infof("payment_adapter||update_order,err=%v", err)
		}
	}()
	ctx, span := tracing.Start(ctx, "order_grpc.update_order")
	defer span.End()

	_, err = o.client.UpdateOrder(ctx, order)
	return status.Convert(err).Err()
}
