package models

import (
	"time"
)

type OrderStatusValue string

var OrderStatus = struct {
	NEW        OrderStatusValue
	PROCESSING OrderStatusValue
	INVALID    OrderStatusValue
	PROCESSED  OrderStatusValue
}{
	NEW:        "NEW",
	PROCESSING: "PROCESSING",
	INVALID:    "INVALID",
	PROCESSED:  "PROCESSED",
}

func (o OrderStatusValue) FromEnum() OrderStatusValue {
	if o != OrderStatus.NEW &&
		o != OrderStatus.PROCESSING &&
		o != OrderStatus.PROCESSED &&
		o != OrderStatus.INVALID {
		return ""
	}
	return o
}

type Order struct {
	ID         int
	UserID     string
	Status     OrderStatusValue
	UploadedAt time.Time
}

type OrderInfo struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
