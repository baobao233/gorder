package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/baobao233/gorder/common/entity"
	"github.com/baobao233/gorder/common/tracing"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type StripeProcessor struct {
	apiKey string
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	if apiKey == "" {
		panic("api key is empty")
	}
	stripe.Key = apiKey
	return &StripeProcessor{apiKey: apiKey}
}

const (
	successURL = "http://localhost:8282/success"
)

func (s StripeProcessor) CreatePaymentLink(ctx context.Context, order *entity.Order) (string, error) {
	// 使用全局的 tracer，不然重新定义 tracer 的话会把同一条链路中的别的 tracer 覆盖掉
	_, span := tracing.Start(ctx, "stripe_processor.create_payment_link")
	defer span.End()

	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}
	marshalledItems, _ := json.Marshal(order.Items)
	metaData := map[string]string{
		"orderID":     order.ID,
		"customerID":  order.CustomerID,
		"status":      order.Status,
		"items":       string(marshalledItems),
		"paymentLink": order.PaymentLink,
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:   metaData,
		LineItems:  items,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("%s?customerID=%s&orderID=%s", successURL, order.CustomerID, order.ID)), // 支付成功后跳转的链接
	}
	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	// 返回第三方支付页面
	return result.URL, nil
}
