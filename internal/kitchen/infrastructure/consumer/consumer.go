package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/convertor"
	"github.com/baobao233/gorder/common/entity"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/common/logging"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"time"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, request *orderpb.Order) error
}

type Consumer struct {
	orderGRPC OrderService
}

func NewConsumer(orderGRPC OrderService) *Consumer {
	return &Consumer{
		orderGRPC: orderGRPC,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	// exclusive设置为 true 是因为假设我们有很多个kitchen 服务，那我们只能消费其中一个消息
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil); err != nil {
		logrus.Fatal(err)
	}

	// 消费消息, msg就是具体的 message
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("fail to consume: queue=%s, err=%s", q.Name, err)
	}

	var forever chan struct{}
	// 开启一个协程处理消息
	go func() {
		for {
			for msg := range msgs {
				c.handleMessage(ch, msg, q)
			}
		}
	}()

	// 阻塞
	<-forever
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers), fmt.Sprintf("rabbitmq.%s.consume", q.Name)) // 抽取 ctx 并开启一个 span
	defer span.End()

	// 有错 nack，无错 ack
	var err error
	defer func() {
		if err != nil {
			logging.Warnf(ctx, nil, "consume failed||from=%s||msg=%+v||err=%v", q.Name, msg, err)
			_ = msg.Nack(false, false) // 回复生产者没有接收到消息
		} else {
			logging.Infof(ctx, nil, "%v", "consume success")
			_ = msg.Ack(false) // 回复生产者接收到消息
		}
	}()

	o := &entity.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		err = errors.Wrap(err, "failed to unmarshall msg to order")
		return
	}

	if o.Status != "paid" {
		err = errors.New("order not paid, can not cook")
		return
	}
	cook(ctx, o)
	span.AddEvent(fmt.Sprintf("order_cook: %v", o))

	if err = c.orderGRPC.UpdateOrder(ctx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      "ready",
		Items:       convertor.NewItemConvertor().EntitiesToProtos(o.Items),
		PaymentLink: o.PaymentLink,
	}); err != nil {
		logging.Errorf(ctx, nil, "error updating order||orderID=%s||err=%v", o.ID, err)
		// Retry
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handle retry, messageID=%s||err=%v", msg.MessageId, err)
		}
		return
	}
	span.AddEvent("kitchen.order.finished.updated")
}

func cook(ctx context.Context, o *entity.Order) {
	logrus.WithContext(ctx).Printf(fmt.Sprintf("cooking order, orderID: %s", o.ID))
	time.Sleep(5 * time.Second)
	logrus.WithContext(ctx).Printf("order %s done!", o.ID)
}
