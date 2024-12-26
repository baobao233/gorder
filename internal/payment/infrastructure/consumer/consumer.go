package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/payment/app"
	"github.com/baobao233/gorder/payment/app/command"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
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
				c.handleMessage(ch, msg, q)
			}
		}
	}()

	// 阻塞
	<-forever
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	logrus.Infof("Payment recieve a message from %s, msg=%s", q.Name, string(msg.Body))
	// extract span
	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	t := otel.Tracer("rabbitmq")
	_, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
	defer span.End()

	// 有错 nack，无错 ack
	var err error
	defer func() {
		if err != nil {
			_ = msg.Nack(false, false) // 回复生产者没有接收到消息
		} else {
			_ = msg.Ack(false) // 回复生产者接收到消息
		}
	}()

	o := &orderpb.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		logrus.Warnf("failed to unmarshall msg to order, err=%v", err)
		return
	}
	_, err = c.app.Command.CreatePayment.Handle(ctx, command.CreatePayment{Order: o})
	if err != nil {
		logrus.Warnf("failed to create payment, err=%v", err)
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			logrus.Warnf("retry_error, error handle retry, messageID=%s, err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("payment.created")
	logrus.Info("consume success")
}
