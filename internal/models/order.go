// Package models содержит описания структур данных (DTO),
// которые используются во всем приложении и для маппинга JSON/DB.
package models

import "time"

// Order представляет полную информацию о заказе, включая данные о доставке,
// оплате и списке приобретенных товаров.
type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required"`
	TrackNumber       string    `json:"track_number" validate:"required"`
	Entry             string    `json:"entry" validate:"required"`
	Delivery          Delivery  `json:"delivery" validate:"required"`
	Payment           Payment   `json:"payment" validate:"required"`
	Items             []Items   `json:"items" validate:"required,gt=0,dive"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shard_key"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

// Delivery содержит контактные данные получателя и адрес доставки заказа.
type Delivery struct {
	Name    string `json:"name" validate:"required,min=2"`
	Phone   string `json:"phone" validate:"required,e164"`
	Zip     string `json:"zip" validate:"required,numeric"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

// Payment описывает финансовые параметры транзакции, включая информацию о банке,
// валюте и итоговой стоимости заказа.
type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id" validate:"required"`
	Currency     string `json:"currency" validate:"required,len=3"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"gt=0"`
	PaymentDt    int    `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"gt=0"`
	CustomFee    int    `json:"custom_fee" validate:"required"`
}

// Items содержит детальную информацию о конкретной позиции (товаре) в заказе.
type Items struct {
	ChrtID      int    `json:"chrt_id" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"gt=0"` // Цена товара > 0
	Rid         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price" validate:"gt=0"`
	NmID        int    `json:"nm_id" validate:"required"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
