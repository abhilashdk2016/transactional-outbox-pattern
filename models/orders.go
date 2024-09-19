package models

type Orders struct {
	ID         uint    `json:"id"`
	CustomerId int     `json:"customer_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}
