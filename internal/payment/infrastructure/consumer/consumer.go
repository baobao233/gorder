package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel"

	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/payment/app"
	"github.com/baobao233/gorder/payment/app/command"
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
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
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
				c.handleMessage(msg, q)
			}
		}
	}()

	// 阻塞
	<-forever
}

func (c *Consumer) handleMessage(msg amqp.Delivery, q amqp.Queue) {
	logrus.Infof("Payment recieve a message from %s, msg=%s", q.Name, string(msg.Body))
	// extract span
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	_, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	o := &orderpb.Order{}
	if err := json.Unmarshal(msg.Body, o); err != nil {
		logrus.Warnf("failed to unmarshall msg to order, err=%v", err)
		_ = msg.Nack(false, false)
		return
	}
	_, err := c.app.Command.CreatePayment.Handle(ctx, command.CreatePayment{Order: o})
	if err != nil {
		// TODO: Retry
		logrus.Warnf("failed to create order, err=%v", err)
		_ = msg.Nack(false, false)
		return
	}

	span.AddEvent("payment.created")
	_ = msg.Ack(false) // 回复生产者有没有接收到消息
	logrus.Info("consume success")
}
