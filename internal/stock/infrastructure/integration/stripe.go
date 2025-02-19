package integration

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81/product"

	_ "github.com/baobao233/gorder/common/config"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v81"
)

type StripeAPI struct {
	stripeKey string
}

func NewStripeAPI() *StripeAPI {
	key := viper.GetString("stripe-key")
	if key == "" {
		logrus.Fatal("empty key")
	}
	return &StripeAPI{stripeKey: key}
}

func (s *StripeAPI) GetPriceByProductID(ctx context.Context, pid string) (string, error) {
	stripe.Key = s.stripeKey
	result, err := product.Get(pid, &stripe.ProductParams{})
	if err != nil {
		return "", err
	}
	return result.DefaultPrice.ID, nil
}

func (s *StripeAPI) GetProductByID(ctx context.Context, pid string) (*stripe.Product, error) {
	stripe.Key = s.stripeKey
	return product.Get(pid, &stripe.ProductParams{})
}
