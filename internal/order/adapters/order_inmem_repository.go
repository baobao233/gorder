package adapters

import (
	"context"
	domain "github.com/baobao233/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

// MemoryOrderRepository Repository的具体实现
type MemoryOrderRepository struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	s := make([]*domain.Order, 0)
	s = append(s, &domain.Order{
		ID:          "fake-id",
		CustomerID:  "fake-customer-id",
		Status:      "fake-status",
		PaymentLink: "fake-payment-link",
		Items:       nil,
	})
	return &MemoryOrderRepository{
		lock:  &sync.RWMutex{},
		store: s,
	}
}

func (m MemoryOrderRepository) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	newOrder := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
	m.store = append(m.store, newOrder)
	logrus.WithFields(logrus.Fields{
		"input_order":         order,
		"store_after_create_": m.store,
	}).Debug("memory_order_repo_create")
	return newOrder, nil
}

func (m MemoryOrderRepository) Get(_ context.Context, id, customerID string) (*domain.Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, order := range m.store {
		if order.ID == id && order.CustomerID == customerID {
			logrus.Debug("memory_order_repo_get || found || id=%s || customerID=%s || res=%v", order.ID, order.CustomerID, *order)
			return order, nil
		}
	}
	return nil, domain.NotFoundError{OrderID: id}
}

func (m MemoryOrderRepository) Update(ctx context.Context, o *domain.Order, updateFunc func(context.Context, *domain.Order) (*domain.Order, error)) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	found := false
	for i, order := range m.store {
		if order.ID == o.ID && order.CustomerID == o.CustomerID {
			found = true
			updateOrder, err := updateFunc(ctx, order)
			if err != nil {
				return err
			}
			m.store[i] = updateOrder
		}
	}
	if !found {
		return domain.NotFoundError{OrderID: o.ID}
	}
	return nil
}
