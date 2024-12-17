package processor

import (
	"context"
	"encoding/json"
	"github.com/baobao233/gorder/common/genproto/orderpb"
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

var (
	successURL = "http://localhost:8282/success"
)

func (s StripeProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {
	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID), // 假设恒定
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}
	marshalledItems, _ := json.Marshal(order.Items)
	metaData := map[string]string{
		"orderID":    order.ID,
		"customerID": order.CustomerID,
		"status":     order.Status,
		"items":      string(marshalledItems),
	}
	params := &stripe.CheckoutSessionParams{
		Metadata:   metaData,
		LineItems:  items,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL), // 支付成功后跳转的链接
	}
	result, err := session.New(params)
	if err != nil {
		return "", err
	}
	// 返回第三方支付页面
	return result.URL, nil
}
