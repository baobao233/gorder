package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/genproto/orderpb"
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

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*orderpb.Item
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
	var err error
	logrus.Infof("Kitchen recieve a message from %s, msg=%s", q.Name, string(msg.Body))
	// extract span
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	mqCtx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))

	// 有错 nack，无错 ack
	defer func() {
		span.End()
		if err != nil {
			_ = msg.Nack(false, false) // 回复生产者没有接收到消息
		} else {
			_ = msg.Ack(false) // 回复生产者接收到消息
		}
	}()

	o := &Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		logrus.Warnf("failed to unmarshall msg to order, err=%v", err)
		return
	}

	if o.Status != "paid" {
		err = errors.New("order not paid, can not cook")
		return
	}
	cook(o)
	span.AddEvent(fmt.Sprintf("order_cook: %v", o))

	if err := c.orderGRPC.UpdateOrder(mqCtx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      "ready",
		Items:       o.Items,
		PaymentLink: o.PaymentLink,
	}); err != nil {
		// Retry
		if err = broker.HandleRetry(mqCtx, ch, &msg); err != nil {
			logrus.Warnf("kitchen: error handling retry, err=%v", err)
		}
		return
	}
	span.AddEvent("kitchen.order.finished.updated")
	logrus.Info("consume success")
}

func cook(o *Order) {
	logrus.Info(fmt.Sprintf("cooking order, orderID: %s", o.ID))
	time.Sleep(5 * time.Second)
	logrus.Info(fmt.Sprintf("order %s done!", o.ID))
}
