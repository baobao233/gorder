package tests

import (
	"context"
	"fmt"
	"github.com/spf13/viper"

	sw "github.com/baobao233/gorder/common/client/order"
	_ "github.com/baobao233/gorder/common/config"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

var (
	ctx    = context.Background()
	server = fmt.Sprintf("http://%s/api", viper.GetString("order.http-addr"))
)

func TestMain(m *testing.M) {
	before()
	m.Run()
}

func before() {
	log.Printf("server=%s", server)
}

func TestCreateOrder_success(t *testing.T) {
	response := getResponse(t, "123", sw.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: "123",
		Items: []sw.ItemWithQuantity{
			{
				Id:       "prod_ROGERol0hYd3tx",
				Quantity: 1,
			},
			{
				Id:       "prod_RPmhK2nexllK3r",
				Quantity: 20,
			},
		},
	})
	t.Logf("body=%s", response.Body)
	assert.Equal(t, 200, response.StatusCode())
	assert.Equal(t, 0, response.JSON200.Errno)
}

func TestCreateOrder_invalidParams(t *testing.T) {
	response := getResponse(t, "123", sw.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: "123",
		Items:      nil,
	})
	t.Logf("body=%s", response.Body)
	assert.Equal(t, 200, response.StatusCode())
	assert.Equal(t, 2, response.JSON200.Errno)
}

func getResponse(t *testing.T, customerID string, body sw.PostCustomerCustomerIdOrdersJSONRequestBody) *sw.PostCustomerCustomerIdOrdersResponse {
	t.Helper()
	client, err := sw.NewClientWithResponses(server)
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.PostCustomerCustomerIdOrdersWithResponse(ctx, customerID, body)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
