package command

import (
	"context"
	"fmt"
	"github.com/baobao233/gorder/common/logging"
	"github.com/baobao233/gorder/order/convertor"
	"github.com/baobao233/gorder/order/entity"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/status"

	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/decorator"
	"github.com/baobao233/gorder/order/app/query"
	domain "github.com/baobao233/gorder/order/domain/order"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type CreateOrder struct {
	CustomerID string
	Items      []*entity.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string // 只暴露一个信息
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderCommand struct {
	orderRepo domain.Repository
	stockGRPC query.StockService // 依赖于接口，而不是直接在 NewCreateOrderHandler 中去直接传参
	channel   *amqp.Channel
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockService,
	channel *amqp.Channel,
	logger *logrus.Entry,
	metricsClient decorator.MetricClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	if stockGRPC == nil {
		panic("nil stockGRPC")
	}
	if channel == nil {
		panic("nil channel")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderCommand{
			orderRepo: orderRepo,
			stockGRPC: stockGRPC,
			channel:   channel,
		},
		logger,
		metricsClient)
}

func (c createOrderCommand) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	var err error
	defer logging.WhenCommandExecute(ctx, "CreateOrderHandler", cmd, err)

	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderCreated)) // Create a span to track validate() and PublishWithContext()
	defer span.End()

	// 调用 stockGRPC 校验
	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
	}

	pendingOrder, err := domain.NewPendingOrder(cmd.CustomerID, validItems)
	if err != nil {
		return nil, err
	}
	o, err := c.orderRepo.Create(ctx, pendingOrder)
	if err != nil {
		return nil, err
	}

	err = broker.PublishEvent(ctx, broker.PublishEventReq{
		Channel:  c.channel,
		Routing:  broker.Direct,
		Exchange: "",
		Queue:    broker.EventOrderCreated,
		Body:     o,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "publish event error, q.Name=%s", broker.EventOrderCreated)
	}

	return &CreateOrderResult{OrderID: o.ID}, nil
}

func (c createOrderCommand) validate(ctx context.Context, items []*entity.ItemWithQuantity) ([]*entity.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have at least 1 item")
	}
	items = packItems(items)                                                                                            // 合并同类型的 item
	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, convertor.NewItemWithQuantityConvertor().EntitiesToProtos(items)) // 进入到 grpc 时候需要转换一层
	if err != nil {
		return nil, status.Convert(err).Err()
	}
	return convertor.NewItemConvertor().ProtosToEntities(resp.Items), nil
}

func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}

	var res []*entity.ItemWithQuantity
	for id, quantity := range merged {
		res = append(res, &entity.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}
	return res
}
