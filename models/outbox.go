package models

type Outbox struct {
	ID          uint   `json:"id"`
	Payload     string `json:"payload"`
	IsProcessed bool   `json:"is_processed"`
}
