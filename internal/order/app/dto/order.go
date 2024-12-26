package dto

type CreateOrderResp struct {
	CustomerID  string `json:"customer_id"`
	OrderID     string `json:"order_id"`
	RedirectURL string `json:"redirect_url"`
}
