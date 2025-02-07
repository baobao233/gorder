package main

import (
	"encoding/json"
	"fmt"
	"github.com/baobao233/gorder/common/logging"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"

	"github.com/baobao233/gorder/common/broker"
	"github.com/baobao233/gorder/common/genproto/orderpb"
	"github.com/baobao233/gorder/payment/domain"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type PaymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(ch *amqp.Channel) *PaymentHandler {
	return &PaymentHandler{channel: ch}
}

// RegisterRoutes cmd: stripe listen --forward-to localhost:8284/api/webhook
func (h *PaymentHandler) RegisterRoutes(c *gin.Engine) {
	// 注册一个路由，用于 stripe 返回结果到的一个 link
	c.POST("/api/webhook", h.handleWebhook)
}

// handleWebhook 处理从 stripe 传回来的信息，stripe 与 payment 之间通过 checkoutSession 传递信息，将订单已经支付的消息发送到 mq 中
func (h *PaymentHandler) handleWebhook(c *gin.Context) {
	logrus.WithContext(c.Request.Context()).Info("receive webhook from stripe")
	var err error
	defer func() {
		if err != nil {
			logging.Warnf(c.Request.Context(), nil, "handleWebhook err=%v", err)
		} else {
			logging.Infof(c.Request.Context(), nil, "%v", "handleWebhook success")
		}
	}()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		err = errors.Wrap(err, "Error reading request body")
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("ENDPOINT_STRIPE_SECRET")) // 看备忘录获得具体来源

	if err != nil {
		err = errors.Wrap(err, "error verifying webhook signature")
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err = json.Unmarshal(event.Data.Raw, &session); err != nil {
			err = errors.Wrap(err, "error unmarshall event.data.raw into session")
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			var items []*orderpb.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)

			if err != nil {
				err = errors.Wrap(err, "error marshall domain order")
				return
			}

			tr := otel.Tracer("rabbitmq")
			ctx, span := tr.Start(c.Request.Context(), fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid))
			defer span.End()

			// 发送给 mq，传递给 order 去 update
			_ = broker.PublishEvent(ctx, broker.PublishEventReq{
				Channel:  h.channel,
				Routing:  broker.FanOut,
				Exchange: broker.EventOrderPaid,
				Queue:    "",
				Body: &domain.Order{
					ID:          session.Metadata["orderID"],
					CustomerID:  session.Metadata["customerID"],
					Status:      string(stripe.CheckoutSessionPaymentStatusPaid),
					PaymentLink: session.Metadata["paymentLink"],
					Items:       items,
				},
			})
		}
	}
	c.JSON(http.StatusOK, nil)
}
