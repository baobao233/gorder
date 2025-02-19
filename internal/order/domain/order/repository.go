package order

import (
	"context"
	"fmt"
)

type Repository interface {
	Create(context.Context, *Order) (*Order, error)
	Get(ctx context.Context, id, customerID string) (*Order, error)
	Update(
		ctx context.Context,
		o *Order,
		updateFunc func(context.Context, *Order) (*Order, error), // 传入一个需要被更新状态的 order， 返回一个更新状态后的 order，更新的动作由调用 GRPC 的服务完成
	) error
}

type NotFoundError struct {
	OrderID string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("order '%s' not found", e.OrderID)
}
