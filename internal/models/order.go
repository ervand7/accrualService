package models

import (
	"time"
)

type OrderStatusEnumValue string

var OrderStatus = struct {
	NEW        OrderStatusEnumValue
	PROCESSING OrderStatusEnumValue
	INVALID    OrderStatusEnumValue
	PROCESSED  OrderStatusEnumValue
}{
	NEW:        "NEW",
	PROCESSING: "PROCESSING",
	INVALID:    "INVALID",
	PROCESSED:  "PROCESSED",
}

type Order struct {
	ID         int
	UserID     string
	Status     OrderStatusEnumValue
	UploadedAt time.Time
}
