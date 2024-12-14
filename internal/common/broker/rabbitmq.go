package broker

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

/*
给 RabbitMQ 做相应的初始化
*/

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

	return ch, conn.Close
}
