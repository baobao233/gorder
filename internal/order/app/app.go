package app

import "github.com/baobao233/gorder/order/app/query"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
}

type Queries struct {
	GetCustomOrder query.GetCustomerOrderHandler
}
