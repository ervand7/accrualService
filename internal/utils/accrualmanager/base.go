package accrualmanager

import (
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	"time"
)

const Timeout time.Duration = 2

var accrualSystemURL = config.GetConfig().AccrualSystemAddress + "/api/orders/"

type OrderStatusType string

var OrderStatus = struct {
	REGISTERED OrderStatusType
	INVALID    OrderStatusType
	PROCESSING OrderStatusType
	PROCESSED  OrderStatusType
}{
	REGISTERED: "REGISTERED",
	INVALID:    "INVALID",
	PROCESSING: "PROCESSING",
	PROCESSED:  "PROCESSED",
}
