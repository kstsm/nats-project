package models

import (
	"github.com/google/uuid"
	"time"
)

type Order struct {
	OrderUID          uuid.UUID `json:"id"`
	TrackNumber       string    `json:"track_number,"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardKey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	ID           int       `json:"id"`            // Уникальный идентификатор платежа
	Transaction  string    `json:"transaction"`   // Идентификатор транзакции
	RequestID    string    `json:"request_id"`    // Идентификатор запроса (может быть пустым)
	Currency     string    `json:"currency"`      // Валюта платежа
	Provider     string    `json:"provider"`      // Платежный провайдер
	Amount       float64   `json:"amount"`        // Сумма платежа
	PaymentDT    time.Time `json:"payment_dt"`    // Дата и время платежа
	Bank         string    `json:"bank"`          // Банк
	DeliveryCost int       `json:"delivery_cost"` // Стоимость доставки
	GoodsTotal   int       `json:"goods_total"`   // Общая стоимость товаров
	CustomFee    int       `json:"custom_fee"`    // Таможенные сборы
}
