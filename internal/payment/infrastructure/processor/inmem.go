package processor

import (
	"context"
	"github.com/baobao233/gorder/common/entity"
)

type InMemProcessor struct {
}

func NewInMemProcessor() *InMemProcessor {
	return &InMemProcessor{}
}

func (i InMemProcessor) CreatePaymentLink(ctx context.Context, order *entity.Order) (string, error) {
	return "inmen-payment-link", nil
}
