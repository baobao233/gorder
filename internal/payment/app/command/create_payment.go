package command

import (
	"context"
	"github.com/baobao233/gorder/common/consts"
	"github.com/baobao233/gorder/common/convertor"
	"github.com/baobao233/gorder/common/decorator"
	"github.com/baobao233/gorder/common/entity"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/payment/domain"
	"github.com/sirupsen/logrus"
)

type CreatePayment struct {
	Order *entity.Order
}

// CreatePaymentHandler 输入是 CreatePayment， 输出是支付链接 string
type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (string, error) {
	var err error
	defer logging.WhenRequest(ctx, "CreatePaymentHandler", cmd, err)

	link, err := c.processor.CreatePaymentLink(ctx, cmd.Order)
	if err != nil {
		return "", err
	}
	// 更新新订单的状态和link
	newOrder, err := entity.NewValidOrder(
		cmd.Order.ID,
		cmd.Order.CustomerID,
		consts.OrderStatusWaitingForPayment,
		link,
		cmd.Order.Items,
	)
	if err != nil {
		return "", err
	}
	err = c.orderGRPC.UpdateOrder(ctx, convertor.NewOrderConvertor().EntityToProto(newOrder))

	return link, err
}

func NewCreatePaymentHandler(processor domain.Processor, orderGRPC OrderService, logger *logrus.Logger, metricsClient decorator.MetricClient) CreatePaymentHandler {
	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{processor: processor, orderGRPC: orderGRPC},
		logger, metricsClient)
}
