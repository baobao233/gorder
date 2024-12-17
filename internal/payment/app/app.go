package app

import "github.com/baobao233/gorder/payment/app/command"

type Application struct {
	Command Commands
}

type Commands struct {
	CreatePayment command.CreatePaymentHandler
}
