package accrualmanager

//import (
//	"encoding/json"
//	"fmt"
//	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
//	"io"
//	"net/http"
//	"strconv"
//	"time"
//)
//
//type accrualManager struct {
//	ch         chan int
//	buf        []int
//	resetTimer bool
//	timer      *time.Timer
//}
//
//func (m accrualManager) CollectOrderData(orderNum int) (success bool, err error) {
//	url := accrualSystemURL + strconv.Itoa(orderNum)
//	response, err := http.Get(url)
//	if err != nil {
//		logger.Logger.Error(err.Error())
//		return false, nil
//	}
//
//	defer response.Body.Close()
//	rawBody, err := io.ReadAll(response.Body)
//	if err != nil {
//		logger.Logger.Error(err.Error())
//		return false, nil
//	}
//	if response.StatusCode == http.StatusTooManyRequests {
//		m.timer.Reset(time.Second * Timeout)
//		logger.Logger.Warn(string(rawBody))
//		return
//	}
//
//	type OrderData struct {
//		Order   string `json:"order"`
//		Status  string `json:"status"`
//		Accrual string `json:"accrual,omitempty"`
//	}
//	var orderData OrderData
//	if err := json.Unmarshal(rawBody, &orderData); err != nil {
//		logger.Logger.Error(err.Error())
//		return
//	}
//
//	fmt.Println(string(rawBody))
//}
