package domain

import (
	"context"
	"github.com/baobao233/gorder/common/genproto/orderpb"
)

type Processor interface {
	CreatePaymentLink(context.Context, *orderpb.Order) (string, error)
}
