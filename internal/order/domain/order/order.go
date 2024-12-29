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

// NewOrder 创建代码中流通的 order，也就是 domain.Order
func NewOrder(id, customerID, status string, paymentLink string, items []*entity.Item) (*Order, error) {
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

// NewPendingOrder 创建用于 mongodb Create 的 order，也就是 domain.Order
func NewPendingOrder(customerID string, items []*entity.Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	// ps: payment可以为空，因为的订单已开始创建的时候就是 paymentLink 为空的
	return &Order{
		CustomerID: customerID,
		Status:     "pending",
		Items:      items,
	}, nil
}

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}
	return fmt.Errorf("order status not paid, order id= %s, order status = %s", o.ID, o.Status)
}
