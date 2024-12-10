package main

import (
	"github.com/baobao233/gorder/order/app"
	"github.com/baobao233/gorder/order/app/query"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) PostCustomerCustomerIDOrders(c *gin.Context, customerID string) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) GetCustomerCustomerIDOrdersOrderID(c *gin.Context, customerID string, orderID string) {
	o, err := H.app.Queries.GetCustomOrder.Handle(c, query.GetCustomerOrder{
		CustomerID: "fake-customer-id",
		OrderID:    "fake-id",
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    o,
	})
}
