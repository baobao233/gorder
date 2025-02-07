package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/baobao233/gorder/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/order/app"
	"github.com/baobao233/gorder/order/app/command"
	domain "github.com/baobao233/gorder/order/domain/order"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{
		app: app,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderPaid, false, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	// 绑定到一个 exchange 上
	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
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
			logging.Warnf(ctx, nil, "%v", "consume success")
			_ = msg.Ack(false) // 回复生产者接收到消息
		}
	}()

	o := &domain.Order{}

	if err = json.Unmarshal(msg.Body, &o); err != nil {
		err = errors.Wrap(err, "failed to unmarshall msg to order")
		return
	}

	_, err = c.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			// 校验是否订单状态已经改变，如果没有改变则返回错误
			if err := order.IsPaid(); err != nil {
				return nil, err
			}
			return order, nil
		},
	})
	if err != nil {
		logging.Errorf(ctx, nil, "error updating order||orderID=%s||err=%v", o.ID, err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			err = errors.Wrapf(err, "retry_error, error handle retry, messageID=%s||err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("order.updated") // 如果没有报错可以添加一个事件
}
