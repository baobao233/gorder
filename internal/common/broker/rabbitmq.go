package broker

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"time"

	_ "github.com/baobao233/gorder/common/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	DLX                = "dlx"
	DLQ                = "dlq"
	amqpRetryHeaderKey = "x-amqp-count"
)

var (
	maxRetryCount = viper.GetInt64("rabbitmq.max-retry")
)

// Connect 给 RabbitMQ 做相应的初始化
func Connect(user, password, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
	conn, err := amqp.Dial(address) // 连接到 address 上
	if err != nil {
		logrus.Fatal(err)
	}
	ch, err := conn.Channel() // 得到 channel
	if err != nil {
		logrus.Fatal(err)
	}
	// 由于对于消息队列来讲，生产者被屏蔽，所以我们专注于生产者把消息发送给哪个 exchange 即可，下面就是针对orderCreate和orderPaid创建两个 exchange, 参数的含义可以参考源代码中了解
	err = ch.ExchangeDeclare(EventOrderCreated, "direct", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	// 使用 fanout 广播订单支付成功的消息
	err = ch.ExchangeDeclare(EventOrderPaid, "fanout", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	// 创建死信队列
	err = createDLX(ch)
	if err != nil {
		logrus.Fatal(err)
	}

	return ch, conn.Close
}

func createDLX(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare("share_queue", true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare(DLX, "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.QueueBind(q.Name, "", DLX, false, nil)
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(DLQ, true, false, false, false, nil)
	return err
}

func HandleRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers == nil {
		d.Headers = amqp.Table{}
	}
	retryCount, ok := d.Headers[amqpRetryHeaderKey].(int64)
	if !ok {
		retryCount = 0
	}
	retryCount++
	d.Headers[amqpRetryHeaderKey] = retryCount

	// 超过最大执行次数时执行放入死信队列逻辑
	if retryCount >= maxRetryCount {
		logrus.Infof("moving messages %s to dlq", d.MessageId)
		return ch.PublishWithContext(ctx, "", DLQ, false, false, amqp.Publishing{
			Headers:      d.Headers,
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         d.Body,
		})
	}

	// 没超过时则把消息从哪来就重新 publish 到哪儿去
	logrus.Infof("retrying message %s, count=%d", d.MessageId, retryCount)
	time.Sleep(time.Second * time.Duration(retryCount)) // 根据重试的次数延长重试的时间
	return ch.PublishWithContext(ctx, d.Exchange, d.RoutingKey, false, false, amqp.Publishing{
		Headers:      d.Headers,
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         d.Body,
	})
}

// RabbitMQHeaderCarrier 为 mq 实现链路追踪，实现 carrier
type RabbitMQHeaderCarrier map[string]interface{}

func (r RabbitMQHeaderCarrier) Get(key string) string {
	value, ok := r[key]
	if !ok {
		return ""
	}
	return value.(string)
}

func (r RabbitMQHeaderCarrier) Set(key string, value string) {
	r[key] = value
}

func (r RabbitMQHeaderCarrier) Keys() []string {
	keys := make([]string, len(r))
	i := 0
	for key := range r {
		keys[i] = key
		i++
	}
	return keys
}

func InjectRabbitMQHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

func ExtractRabbitMQHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headers))
}
