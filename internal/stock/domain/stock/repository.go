package stock

import (
	"context"
	"fmt"
	"strings"

	"github.com/baobao233/gorder/stock/entity"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*entity.Item, error)
	GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error)
	UpdateStock(
		ctx context.Context,
		data []*entity.ItemWithQuantity,
		updateFn func(
			ctx context.Context,
			existing []*entity.ItemWithQuantity,
			query []*entity.ItemWithQuantity,
		) ([]*entity.ItemWithQuantity, error),
	) error
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("these items not found in stock: %s", strings.Join(e.Missing, ","))
}

type ExceedStockError struct {
	FailedOn []struct {
		ID   string
		Want int32
		Have int32
	}
}

func (e ExceedStockError) Error() string {
	var info []string
	for _, failed := range e.FailedOn {
		info = append(info, fmt.Sprintf("product_id=%s, want %d, have %d", failed.ID, failed.Want, failed.Have))
	}
	return fmt.Sprintf("not enough stock for [%s]", strings.Join(info, ","))
}
