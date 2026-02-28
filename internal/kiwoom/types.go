package kiwoom

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StockOrderSide indicates buy/sell for order placement.
type StockOrderSide string

const (
	StockOrderSideBuy  StockOrderSide = "buy"
	StockOrderSideSell StockOrderSide = "sell"
)

// DomesticQuote is a typed subset of ka10001 response.
type DomesticQuote struct {
	Symbol     string
	Name       string
	Price      float64
	Open       float64
	High       float64
	Low        float64
	BasePrice  float64
	UpperLimit float64
	LowerLimit float64
	Change     float64
	ChangeRate float64
	Volume     int64
	ReturnMsg  string
	ReturnCode int
}

// ChartCandle is a normalized candle row from ka10081/ka10082/ka10083.
type ChartCandle struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// AccountBalance is a typed subset of kt00005 response.
type AccountBalance struct {
	Deposit              float64
	DepositD1            float64
	DepositD2            float64
	OrderableAmount      float64
	WithdrawableAmount   float64
	UnsettledStockAmount float64
	StockBuyTotalAmount  float64
	EvaluationTotal      float64
	TotalProfitLoss      float64
	TotalProfitLossRate  float64
	PresumedAssetAmount  float64
	CreditLoanTotal      float64
	ReturnMsg            string
	ReturnCode           int
}

// AccountPosition is a typed position row from kt00018.
type AccountPosition struct {
	StockCode        string
	StockName        string
	RemainingQty     int64
	TradableQty      int64
	TodayBuyQty      int64
	TodaySellQty     int64
	PurchasePrice    float64
	CurrentPrice     float64
	PurchaseAmount   float64
	EvaluationAmount float64
	EvaluationProfit float64
	ProfitRate       float64
	WeightRate       float64
	CreditLoanDate   string
}

// UnsettledOrder is a typed row from ka10075.
type UnsettledOrder struct {
	OrderNumber    string
	StockCode      string
	OrderStatus    string
	OrderQty       int64
	UnsettledQty   int64
	OrderPrice     float64
	ConcludedPrice float64
	OrderSideText  string
	ExchangeCode   string
	ExchangeText   string
	ReturnMsg      string
	ReturnCode     int
}

// OrderExecution is a typed row from ka10076.
type OrderExecution struct {
	OrderNumber    string
	StockCode      string
	OrderSideText  string
	ExecutionPrice float64
	ExecutionQty   int64
	OrderTime      string
	OrderStatus    string
	ExchangeCode   string
	ExchangeText   string
}

// InstrumentInfo is a typed subset of ka10100 response.
type InstrumentInfo struct {
	Code       string
	Name       string
	ListCount  int64
	RegDay     string
	State      string
	MarketCode string
	MarketName string
	SectorName string
	ReturnMsg  string
	ReturnCode int
}

// OrderAck represents order ack payload from kt10000/1/2/3.
type OrderAck struct {
	OrderNumber string
	ReturnMsg   string
	ReturnCode  int
}

// PlaceStockOrderRequest is input for kt10000/kt10001.
type PlaceStockOrderRequest struct {
	Side           StockOrderSide
	Exchange       string
	Symbol         string
	Quantity       int64
	OrderPrice     string
	TradeType      string
	ConditionPrice string
}

// CancelStockOrderRequest is input for kt10003.
type CancelStockOrderRequest struct {
	Exchange   string
	OriginalID string
	Symbol     string
	CancelQty  int64
}

// ModifyStockOrderRequest is input for kt10002.
type ModifyStockOrderRequest struct {
	Exchange       string
	OriginalID     string
	Symbol         string
	ModifyQty      int64
	ModifyPrice    string
	ConditionPrice string
}

func asString(v interface{}) string {
	s := strings.TrimSpace(toString(v))
	if s == "<nil>" {
		return ""
	}
	return s
}

func asFloat64(v interface{}) float64 {
	s := asString(v)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "+")
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func asInt64(v interface{}) int64 {
	s := asString(v)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "+")
	if s == "" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return n
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f)
}

func normalizeSymbolCode(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	return strings.TrimPrefix(s, "A")
}

func firstObjectArray(body map[string]interface{}, keys ...string) []map[string]interface{} {
	for _, key := range keys {
		if rows, ok := asObjectArray(body[key]); ok {
			return rows
		}
	}
	for _, v := range body {
		if rows, ok := asObjectArray(v); ok {
			return rows
		}
	}
	return nil
}

func asObjectArray(v interface{}) ([]map[string]interface{}, bool) {
	raw, ok := v.([]interface{})
	if !ok {
		return nil, false
	}
	rows := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		rows = append(rows, row)
	}
	return rows, len(rows) > 0
}

func parseDateYYYYMMDD(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	t, err := time.Parse("20060102", value)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
