package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/jackc/pgconn"
)

const (
	ctxSecond               = 3 * time.Second
	waitDuration            = 1 * time.Second
	threadsCount            = 5
	accrualRegisteredStatus = "REGISTERED"
)

type respBody struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Worker struct {
	storage *database.Storage
}

func StartWorker(storage *database.Storage) {
	worker := &Worker{
		storage: storage,
	}
	go worker.Run()
}

func (w Worker) Run() {
	w.waitBeforeStart(waitDuration)
	for {
		ch := make(chan int)
		var wg sync.WaitGroup
		w.readFromChan(ch, &wg)

		ctx, cancel := context.WithTimeout(context.Background(), ctxSecond)
		orders, err := w.storage.FindOrdersToAccrual(ctx)
		cancel()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				continue
			}
			logger.Logger.Fatal(err.Error())
		}
		for _, val := range orders {
			ch <- val
		}

		close(ch)
		wg.Wait()
	}
}

func (w Worker) readFromChan(ch <-chan int, wg *sync.WaitGroup) {
	for i := 0; i < threadsCount; i++ {
		wg.Add(1)
		go func(ch <-chan int) {
			value := <-ch
			if value != 0 {
				w.updateOrder(value)
			}
			wg.Done()
		}(ch)
	}
}

func (w Worker) updateOrder(orderID int) {
	resp := requestAccrualServer(orderID)
	id := w.prepareID(resp.Order)
	status := w.prepareStatus(resp.Status)
	accrual := resp.Accrual

	ctx, cancel := context.WithTimeout(context.Background(), ctxSecond)
	defer cancel()
	err := w.storage.UpdateOrderAndAccrual(ctx, id, status, accrual)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
}

func (w Worker) waitBeforeStart(delta time.Duration) {
	time.Sleep(delta)
}

func (w Worker) prepareID(ID string) int {
	result, err := strconv.Atoi(ID)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
	return result
}

func (w Worker) prepareStatus(status string) d.OrderStatusValue {
	if status == accrualRegisteredStatus {
		status = string(d.OrderStatus.NEW)
	}
	return d.OrderStatusValue(status).FromEnum()
}

var requestAccrualServer = func(orderID int) respBody {
	for {
		method := fmt.Sprintf("/api/orders/%d", orderID)
		url := config.GetConfig().AccrualSystemAddress + method
		resp, err := http.Get(url)
		if err != nil {
			logger.Logger.Fatal(err.Error())
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			time.Sleep(waitDuration)
			logger.Logger.Warn(
				fmt.Sprintf("order_id: %d resp %s", orderID, resp.Status))
			continue
		}

		var body respBody
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
