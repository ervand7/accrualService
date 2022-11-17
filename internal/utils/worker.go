package utils

import (
	"encoding/json"
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type accrualServerRespBody struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Worker struct {
	storage         *database.Storage
	goroutinesCount int
}

func StartWorker(storage *database.Storage) {
	worker := &Worker{
		storage:         storage,
		goroutinesCount: 5,
	}
	go worker.Run()
}

func (w Worker) Run() {
	w.waitCreateDB()
	var lastID interface{}
	for {
		ch := make(chan int)
		var wg sync.WaitGroup
		for i := 0; i < w.goroutinesCount; i++ {
			wg.Add(1)
			go func(ch <-chan int) {
				value := <-ch
				if value != 0 {
					w.actualizeOrder(value)
				}
				wg.Done()
			}(ch)
		}

		orders, err := w.storage.FindOrdersToAccrual(lastID)
		if err != nil {
			logger.Logger.Fatal(err.Error())
		}
		if orders == nil {
			lastID = nil
		}
		for _, val := range orders {
			lastID = val
			ch <- val
		}

		close(ch)
		wg.Wait()
	}
}

func (w Worker) actualizeOrder(orderID int) {
	resp := w.requestAccrualServer(orderID)
	id := w.prepareID(resp.Order)
	status := w.prepareStatus(resp.Status)
	accrual := resp.Accrual
	err := w.storage.UpdateOrderAndAccrual(id, status, accrual)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
}

func (w Worker) requestAccrualServer(orderID int) accrualServerRespBody {
	for {
		method := fmt.Sprintf("/api/orders/%d", orderID)
		url := config.GetConfig().AccrualSystemAddress + method
		resp, err := http.Get(url)
		if err != nil {
			logger.Logger.Fatal(err.Error())
		}
		if resp.StatusCode != http.StatusOK {
			time.Sleep(time.Second)
			logger.Logger.Warn(
				fmt.Sprintf("order_id: %d resp %s", orderID, resp.Status))
			continue
		}

		var body accrualServerRespBody
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Logger.Error(err.Error())
		}
		if err = json.Unmarshal(data, &body); err != nil {
			logger.Logger.Error(err.Error())
		}

		resp.Body.Close()
		return body
	}
}

func (w Worker) waitCreateDB() {
	time.Sleep(time.Second)
}

func (w Worker) prepareID(ID string) int {
	result, err := strconv.Atoi(ID)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
	return result
}

func (w Worker) prepareStatus(status string) database.OrderStatusValue {
	if status == "REGISTERED" {
		status = "NEW"
	}
	return database.OrderStatusValue(status).FromEnum()
}
