package order

import (
	"fmt"
	"github.com/baobao233/gorder/order/entity"

	"github.com/pkg/errors"
	"github.com/stripe/stripe-go/v81"
)

// Order Aggregate
type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}

// NewOrder 将 orderpb 中的 order 转化成在代码中流通的 order
func NewOrder(id, customerID, status, paymentLink string, items []*entity.Item) (*Order, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}
	if status == "" {
		return nil, errors.New("empty status")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	// ps: payment可以为空，因为的订单已开始创建的时候就是 paymentLink 为空的
	return &Order{
		ID:          id,
		CustomerID:  customerID,
		Status:      status,
		PaymentLink: paymentLink,
		Items:       items,
	}, nil
}

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}
	return fmt.Errorf("order status not paid, order id= %s, order status = %s", o.ID, o.Status)
}
