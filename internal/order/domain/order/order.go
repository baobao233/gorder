package order

import (
	"fmt"
	"github.com/baobao233/gorder/common/consts"
	"github.com/baobao233/gorder/common/entity"
	"github.com/pkg/errors"
	"slices"
)

// Order Aggregate
type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}

func (o *Order) UpdateOrderStatus(to string) error {
	if !o.isValidStatusTransaction(to) {
		return fmt.Errorf("cannot tramsit from %s to %s", o.Status, to)
	}
	o.Status = to
	return nil
}

func (o *Order) UpdatePaymentLink(paymentLink string) error {
	//if paymentLink == "" {
	//	return errors.New("cannot update empty paymentLink")
	//}
	o.PaymentLink = paymentLink
	return nil
}

func (o *Order) UpdateItems(items []*entity.Item) error {
	o.Items = items
	return nil
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
		Status:     consts.OrderStatusPending,
		Items:      items,
	}, nil
}

func (o *Order) isValidStatusTransaction(to string) bool {
	switch o.Status {
	case consts.OrderStatusPending:
		return slices.Contains([]string{consts.OrderStatusWaitingForPayment}, to)
	case consts.OrderStatusWaitingForPayment:
		return slices.Contains([]string{consts.OrderStatusPaid}, to)
	case consts.OrderStatusPaid:
		return slices.Contains([]string{consts.OrderStatusReady}, to)
	default:
		return false
	}
}
