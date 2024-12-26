package main

import (
	"fmt"
	"github.com/baobao233/gorder/common"
	client "github.com/baobao233/gorder/common/client/order"
	"github.com/baobao233/gorder/order/app"
	"github.com/baobao233/gorder/order/app/command"
	"github.com/baobao233/gorder/order/app/dto"
	"github.com/baobao233/gorder/order/app/query"
	"github.com/baobao233/gorder/order/convertor"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	// 定义 request 和 response
	var (
		req  client.CreateOrderRequest
		err  error
		resp dto.CreateOrderResp
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBindJSON(&req); err != nil {
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{ // 因为使用了 otelgin 中间件，所以可以直接使用请求带来的 context 嵌入链路追踪的 span
		CustomerID: req.CustomerId,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items), // 流转在代码内部，需要进行一层转换
	})
	if err != nil {
		return
	}
	resp = dto.CreateOrderResp{
		CustomerID:  req.CustomerId,
		OrderID:     fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerId, r.OrderID),
		RedirectURL: r.OrderID,
	}
}

func (H HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerID string, orderID string) {
	// 定义 response
	var (
		err  error
		resp interface{}
		//resp struct {
		//	Order *client.Order // 因为 client 中的 Order 有 json，所以为了统一 response 中都是 json 格式，所以使用 client.Order
		//}
	)
	defer func() {
		H.Response(c, err, resp)
	}()

	o, err := H.app.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	if err != nil {
		return
	}

	resp = convertor.NewOrderConvertor().EntityToClient(o)
}
