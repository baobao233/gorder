package processor

import (
	"context"

	"github.com/baobao233/gorder/common/genproto/orderpb"
)

type InMemProcessor struct {
}

func NewInMemProcessor() *InMemProcessor {
	return &InMemProcessor{}
}

func (i InMemProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {
	return "inmen-payment-link", nil
}
