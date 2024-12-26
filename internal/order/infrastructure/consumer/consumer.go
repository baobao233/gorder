package consumer

import (
	"context"
	"encoding/json"
	"fmt"
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
	// 抽取 ctx 并开启一个 span
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	_, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	// 有错 nack，无错 ack
	var err error
	defer func() {
		if err != nil {
			_ = msg.Nack(false, false)
		} else {
			_ = msg.Ack(false)
		}
	}()

	o := &domain.Order{}

	if err = json.Unmarshal(msg.Body, &o); err != nil {
		logrus.Infof("error unmarshall msg.body into domain.order, err = %v", err)
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
		logrus.Infof("error updating order, order id = %s, err = %v", o.ID, err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			logrus.Warnf("retry_error, error handle retry, messageID=%s, err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("order.updated") // 如果没有报错可以添加一个事件
	logrus.Infof("order consume paid event success!!")
}
