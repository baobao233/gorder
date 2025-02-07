package command

import (
	"context"
	"github.com/baobao233/gorder/common/logging"

	"github.com/baobao233/gorder/common/decorator"
	domain "github.com/baobao233/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type UpdateOrder struct {
	Order    *domain.Order
	UpdateFn func(context.Context, *domain.Order) (*domain.Order, error)
}

type UpdateOrderHandler decorator.CommandHandler[UpdateOrder, interface{}]

type updateOrderCommand struct {
	orderRepo domain.Repository
	// stockGRPC
}

func NewUpdateOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricClient,
) UpdateOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	return decorator.ApplyCommandDecorators[UpdateOrder, interface{}](
		updateOrderCommand{orderRepo: orderRepo},
		logger,
		metricsClient)
}

func (c updateOrderCommand) Handle(ctx context.Context, cmd UpdateOrder) (interface{}, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "UpdateOrderHandler", cmd, err)

	if cmd.UpdateFn == nil {
		logrus.Panicf("UpdateOrderHandler got nil order, cmd=%+v", cmd)
	}
	if err = c.orderRepo.Update(ctx, cmd.Order, cmd.UpdateFn); err != nil {
		return nil, err
	}
	return nil, nil
}
