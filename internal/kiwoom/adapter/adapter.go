package adapter

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smallfish06/korea-securities-api/internal/kiwoom"
	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// Adapter adapts Kiwoom APIs into broker.Broker.
type Adapter struct {
	client    *kiwoom.Client
	accountID string
	sandbox   bool
	orderDir  string

	mu     sync.RWMutex
	orders map[string]orderContext
}

// Options configures adapter internals.
type Options struct {
	TokenManager    kiwoom.TokenManager
	OrderContextDir string
}

type orderContext struct {
	OrderID      string             `json:"order_id"`
	Symbol       string             `json:"symbol"`
	Exchange     string             `json:"exchange"`
	Side         broker.OrderSide   `json:"side"`
	Quantity     int64              `json:"quantity"`
	RemainingQty int64              `json:"remaining_qty"`
	Price        float64            `json:"price"`
	Status       broker.OrderStatus `json:"status"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// NewAdapterWithOptions creates a Kiwoom adapter with injectable internals.
func NewAdapterWithOptions(sandbox bool, accountID string, opts Options) *Adapter {
	a := &Adapter{
		client:    kiwoom.NewClientWithTokenManager(sandbox, opts.TokenManager),
		accountID: strings.TrimSpace(accountID),
		sandbox:   sandbox,
		orderDir:  strings.TrimSpace(opts.OrderContextDir),
		orders:    make(map[string]orderContext),
	}
	_ = a.loadOrderContexts()
	return a
}

// Name returns broker name.
func (a *Adapter) Name() string {
	return "KIWOOM"
}

// Authenticate authenticates with Kiwoom.
func (a *Adapter) Authenticate(ctx context.Context, creds broker.Credentials) (*broker.Token, error) {
	return a.client.Authenticate(ctx, creds)
}

// GetQuote retrieves quote for domestic markets.
func (a *Adapter) GetQuote(ctx context.Context, market, symbol string) (*broker.Quote, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	if _, err := toKiwoomExchange(market); err != nil {
		return nil, err
	}

	quote, err := a.client.GetDomesticQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	price := normalizedPrice(quote.Price)
	prevClose := normalizedPrice(quote.BasePrice)
	if prevClose == 0 && (price != 0 || quote.Change != 0) {
		prevClose = price - quote.Change
	}

	return &broker.Quote{
		Symbol:     symbol,
		Market:     normalizeOutputMarket(market),
		Price:      price,
		Open:       normalizedPrice(quote.Open),
		High:       normalizedPrice(quote.High),
		Low:        normalizedPrice(quote.Low),
		Close:      price,
		PrevClose:  prevClose,
		Change:     quote.Change,
		ChangeRate: quote.ChangeRate,
		Volume:     quote.Volume,
		UpperLimit: normalizedPrice(quote.UpperLimit),
		LowerLimit: normalizedPrice(quote.LowerLimit),
		Timestamp:  time.Now(),
	}, nil
}

// GetOHLCV retrieves domestic OHLCV for day/week/month intervals.
func (a *Adapter) GetOHLCV(ctx context.Context, market, symbol string, opts broker.OHLCVOpts) ([]broker.OHLCV, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	if _, err := toKiwoomExchange(market); err != nil {
		return nil, err
	}

	interval := strings.ToLower(strings.TrimSpace(opts.Interval))
	if interval == "" {
		interval = "1d"
	}

	baseDate := time.Now().Format("20060102")
	if !opts.To.IsZero() {
		baseDate = opts.To.Format("20060102")
	}

	var (
		candles []kiwoom.ChartCandle
		err     error
	)

	switch interval {
	case "1d", "d", "day", "daily":
		candles, err = a.client.GetDailyChart(ctx, symbol, baseDate)
	case "1w", "w", "week", "weekly":
		candles, err = a.client.GetWeeklyChart(ctx, symbol, baseDate)
	case "1mo", "mo", "month", "monthly":
		candles, err = a.client.GetMonthlyChart(ctx, symbol, baseDate)
	default:
		return nil, fmt.Errorf("unsupported interval for kiwoom: %s", opts.Interval)
	}
	if err != nil {
		return nil, err
	}
	if len(candles) == 0 {
		return []broker.OHLCV{}, nil
	}

	out := make([]broker.OHLCV, 0, len(candles))
	for _, candle := range candles {
		item := broker.OHLCV{
			Timestamp: candle.Date,
			Open:      normalizedPrice(candle.Open),
			High:      normalizedPrice(candle.High),
			Low:       normalizedPrice(candle.Low),
			Close:     normalizedPrice(candle.Close),
			Volume:    candle.Volume,
		}
		if !opts.From.IsZero() && item.Timestamp.Before(startOfDay(opts.From)) {
			continue
		}
		if !opts.To.IsZero() && item.Timestamp.After(endOfDay(opts.To)) {
			continue
		}
		out = append(out, item)
	}

	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[:opts.Limit]
	}
	return out, nil
}

// GetBalance retrieves account balance summary.
func (a *Adapter) GetBalance(ctx context.Context, accountID string) (*broker.Balance, error) {
	if strings.TrimSpace(accountID) == "" {
		accountID = a.accountID
	}

	bal, err := a.client.GetAccountBalance(ctx, "KRX")
	if err != nil {
		return nil, err
	}

	totalAssets := bal.PresumedAssetAmount
	if totalAssets == 0 {
		totalAssets = bal.Deposit + bal.EvaluationTotal
	}

	return &broker.Balance{
		AccountID:        strings.TrimSpace(accountID),
		Cash:             bal.Deposit,
		TotalAssets:      totalAssets,
		BuyingPower:      bal.OrderableAmount,
		WithdrawableCash: bal.WithdrawableAmount,
		ReceivableAmount: bal.DepositD2,
		ProfitLoss:       bal.TotalProfitLoss,
		ProfitLossPct:    bal.TotalProfitLossRate,
		PositionCost:     bal.StockBuyTotalAmount,
		PositionValue:    bal.EvaluationTotal,
		SettlementT1:     bal.DepositD1,
		Unsettled:        bal.UnsettledStockAmount,
		LoanBalance:      bal.CreditLoanTotal,
	}, nil
}

// GetPositions retrieves account stock positions.
func (a *Adapter) GetPositions(ctx context.Context, _ string) ([]broker.Position, error) {
	positionsResp, err := a.client.GetAccountPositions(ctx, "0", "KRX")
	if err != nil {
		return nil, err
	}
	if len(positionsResp) == 0 {
		return []broker.Position{}, nil
	}

	positions := make([]broker.Position, 0, len(positionsResp))
	for _, row := range positionsResp {
		symbol := normalizeSymbol(row.StockCode)
		if symbol == "" {
			continue
		}
		if row.RemainingQty == 0 {
			continue
		}

		positions = append(positions, broker.Position{
			Symbol:        symbol,
			Name:          row.StockName,
			Market:        "KRX",
			MarketCode:    "KRX",
			AssetType:     broker.AssetStock,
			Quantity:      row.RemainingQty,
			OrderableQty:  row.TradableQty,
			TodayBuyQty:   row.TodayBuyQty,
			TodaySellQty:  row.TodaySellQty,
			AvgPrice:      normalizedPrice(row.PurchasePrice),
			CurrentPrice:  normalizedPrice(row.CurrentPrice),
			PurchaseValue: normalizedPrice(row.PurchaseAmount),
			MarketValue:   normalizedPrice(row.EvaluationAmount),
			ProfitLoss:    row.EvaluationProfit,
			ProfitLossPct: row.ProfitRate,
			WeightPct:     row.WeightRate,
			LoanDate:      row.CreditLoanDate,
		})
	}
	return positions, nil
}

// PlaceOrder places buy/sell stock orders.
func (a *Adapter) PlaceOrder(ctx context.Context, req broker.OrderRequest) (*broker.OrderResult, error) {
	symbol := normalizeSymbol(req.Symbol)
	if symbol == "" || req.Quantity <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}
	exchange, err := toKiwoomExchange(req.Market)
	if err != nil {
		return nil, err
	}

	tradeType, orderPrice, err := toTradeTypeAndPrice(req.Type, req.Price)
	if err != nil {
		return nil, err
	}

	side := kiwoom.StockOrderSideBuy
	if req.Side == broker.OrderSideSell {
		side = kiwoom.StockOrderSideSell
	}

	ack, err := a.client.PlaceStockOrder(ctx, kiwoom.PlaceStockOrderRequest{
		Side:           side,
		Exchange:       exchange,
		Symbol:         symbol,
		Quantity:       req.Quantity,
		OrderPrice:     orderPrice,
		TradeType:      tradeType,
		ConditionPrice: "",
	})
	if err != nil {
		return nil, err
	}

	orderID := strings.TrimSpace(ack.OrderNumber)
	if orderID == "" {
		return nil, fmt.Errorf("missing order id in kiwoom response")
	}

	a.storeOrderContext(orderID, orderContext{
		OrderID:      orderID,
		Symbol:       symbol,
		Exchange:     exchange,
		Side:         req.Side,
		Quantity:     req.Quantity,
		RemainingQty: req.Quantity,
		Price:        req.Price,
		Status:       broker.OrderStatusPending,
		UpdatedAt:    time.Now(),
	})

	return &broker.OrderResult{
		OrderID:      orderID,
		Status:       broker.OrderStatusPending,
		RemainingQty: req.Quantity,
		Message:      strings.TrimSpace(ack.ReturnMsg),
		Timestamp:    time.Now(),
	}, nil
}

// CancelOrder cancels a pending order.
func (a *Adapter) CancelOrder(ctx context.Context, orderID string) error {
	meta, err := a.resolveOrderContext(ctx, orderID)
	if err != nil {
		return err
	}
	cancelQty := meta.RemainingQty
	if cancelQty <= 0 {
		cancelQty = meta.Quantity
	}
	if cancelQty <= 0 {
		cancelQty = 1
	}

	ack, err := a.client.CancelStockOrder(ctx, kiwoom.CancelStockOrderRequest{
		Exchange:   meta.Exchange,
		OriginalID: meta.OrderID,
		Symbol:     meta.Symbol,
		CancelQty:  cancelQty,
	})
	if err != nil {
		return err
	}

	meta.Status = broker.OrderStatusCancelled
	meta.RemainingQty = 0
	meta.UpdatedAt = time.Now()
	a.storeOrderContext(meta.OrderID, meta)

	newID := strings.TrimSpace(ack.OrderNumber)
	if newID != "" && newID != meta.OrderID {
		meta.OrderID = newID
		a.storeOrderContext(newID, meta)
	}
	return nil
}

// ModifyOrder modifies an existing order.
func (a *Adapter) ModifyOrder(ctx context.Context, orderID string, req broker.ModifyOrderRequest) (*broker.OrderResult, error) {
	meta, err := a.resolveOrderContext(ctx, orderID)
	if err != nil {
		return nil, err
	}

	newQty := meta.Quantity
	if req.Quantity > 0 {
		newQty = req.Quantity
	}
	newPrice := meta.Price
	if req.Price > 0 {
		newPrice = req.Price
	}
	if newQty <= 0 || newPrice <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}

	ack, err := a.client.ModifyStockOrder(ctx, kiwoom.ModifyStockOrderRequest{
		Exchange:       meta.Exchange,
		OriginalID:     meta.OrderID,
		Symbol:         meta.Symbol,
		ModifyQty:      newQty,
		ModifyPrice:    formatPrice(newPrice),
		ConditionPrice: "",
	})
	if err != nil {
		return nil, err
	}

	newOrderID := strings.TrimSpace(ack.OrderNumber)
	if newOrderID == "" {
		newOrderID = orderID
	}

	meta.OrderID = newOrderID
	meta.Quantity = newQty
	meta.RemainingQty = newQty
	meta.Price = newPrice
	meta.Status = broker.OrderStatusPending
	meta.UpdatedAt = time.Now()
	a.storeOrderContext(newOrderID, meta)
	if newOrderID != orderID {
		a.storeOrderContext(orderID, meta)
	}

	return &broker.OrderResult{
		OrderID:      newOrderID,
		Status:       broker.OrderStatusPending,
		RemainingQty: newQty,
		Message:      strings.TrimSpace(ack.ReturnMsg),
		Timestamp:    time.Now(),
	}, nil
}

// GetOrder returns order status by order id.
func (a *Adapter) GetOrder(ctx context.Context, orderID string) (*broker.OrderResult, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, broker.ErrInvalidOrderRequest
	}

	meta, _ := a.getOrderContext(orderID)
	unsettled, err := a.fetchUnsettled(ctx, meta.Symbol)
	if err == nil {
		for _, row := range unsettled {
			if strings.TrimSpace(row.OrderNumber) != strings.TrimSpace(orderID) {
				continue
			}
			ordQty := row.OrderQty
			remaining := row.UnsettledQty
			filled := ordQty - remaining
			if filled < 0 {
				filled = 0
			}
			status := mapOrderStatus(row.OrderStatus, remaining)
			result := &broker.OrderResult{
				OrderID:        orderID,
				Status:         status,
				FilledQuantity: filled,
				RemainingQty:   remaining,
				AvgFilledPrice: normalizedPrice(row.ConcludedPrice),
				Message:        row.OrderStatus,
				Timestamp:      time.Now(),
			}
			meta.OrderID = orderID
			meta.Symbol = normalizeSymbol(row.StockCode)
			meta.Exchange = exchangeFromUnsettledRow(row)
			meta.Quantity = ordQty
			meta.RemainingQty = remaining
			meta.Price = normalizedPrice(row.OrderPrice)
			meta.Side = sideFromKiwoomOrderText(row.OrderSideText)
			meta.Status = status
			meta.UpdatedAt = time.Now()
			a.storeOrderContext(orderID, meta)
			return result, nil
		}
	}

	fills, fillErr := a.GetOrderFills(ctx, orderID)
	if fillErr == nil && len(fills) > 0 {
		var qty int64
		var amount float64
		for _, f := range fills {
			qty += f.Quantity
			amount += float64(f.Quantity) * f.Price
		}
		avg := 0.0
		if qty > 0 {
			avg = amount / float64(qty)
		}
		return &broker.OrderResult{
			OrderID:        orderID,
			Status:         broker.OrderStatusFilled,
			FilledQuantity: qty,
			RemainingQty:   0,
			AvgFilledPrice: avg,
			Timestamp:      time.Now(),
		}, nil
	}

	if meta.OrderID != "" {
		return &broker.OrderResult{
			OrderID:      orderID,
			Status:       meta.Status,
			RemainingQty: meta.RemainingQty,
			Timestamp:    time.Now(),
		}, nil
	}
	return nil, broker.ErrOrderNotFound
}

// GetOrderFills returns fills for the order.
func (a *Adapter) GetOrderFills(ctx context.Context, orderID string) ([]broker.OrderFill, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, broker.ErrInvalidOrderRequest
	}
	meta, hasContext := a.getOrderContext(orderID)

	executions, err := a.client.GetOrderExecutions(ctx, meta.Symbol)
	if err != nil {
		return nil, err
	}
	if len(executions) == 0 {
		if hasContext {
			return []broker.OrderFill{}, nil
		}
		return nil, broker.ErrOrderNotFound
	}

	fills := make([]broker.OrderFill, 0)
	for _, row := range executions {
		if strings.TrimSpace(row.OrderNumber) != strings.TrimSpace(orderID) {
			continue
		}
		qty := row.ExecutionQty
		if qty <= 0 {
			continue
		}
		price := normalizedPrice(row.ExecutionPrice)
		fills = append(fills, broker.OrderFill{
			OrderID:   orderID,
			Symbol:    normalizeSymbol(row.StockCode),
			Market:    mapStexCodeToMarket(row.ExchangeCode, row.ExchangeText),
			Side:      sideFromKiwoomText(row.OrderSideText),
			Quantity:  qty,
			Price:     price,
			Amount:    float64(qty) * price,
			Currency:  "KRW",
			FilledAt:  parseOrderTime(row.OrderTime),
			RawStatus: strings.TrimSpace(row.OrderStatus),
		})
	}

	if len(fills) == 0 {
		if hasContext {
			return []broker.OrderFill{}, nil
		}
		return nil, broker.ErrOrderNotFound
	}
	sort.Slice(fills, func(i, j int) bool {
		return fills[i].FilledAt.Before(fills[j].FilledAt)
	})
	return fills, nil
}

// GetInstrument retrieves normalized instrument metadata.
func (a *Adapter) GetInstrument(ctx context.Context, market, symbol string) (*broker.Instrument, error) {
	symbol = normalizeSymbol(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}

	info, err := a.client.GetInstrumentInfo(ctx, symbol)
	if err != nil {
		return nil, err
	}

	marketOut := normalizeOutputMarket(market)
	if marketOut == "" {
		marketOut = marketFromCode(info.MarketCode, info.MarketName)
	}

	state := strings.TrimSpace(info.State)
	listed := !strings.Contains(state, "상장폐지")
	suspended := strings.Contains(state, "거래정지") || strings.Contains(state, "정리매매")

	return &broker.Instrument{
		Symbol:       normalizeSymbol(info.Code),
		Market:       marketOut,
		Name:         firstNonEmpty(info.Name, symbol),
		Exchange:     firstNonEmpty(info.MarketName, marketOut),
		Currency:     "KRW",
		Country:      "KR",
		AssetType:    broker.AssetStock,
		Sector:       info.SectorName,
		ListedShares: info.ListCount,
		IsListed:     listed,
		IsSuspended:  suspended,
		ListingDate:  info.RegDay,
	}, nil
}

func (a *Adapter) fetchUnsettled(ctx context.Context, symbol string) ([]kiwoom.UnsettledOrder, error) {
	return a.client.GetUnsettledOrders(ctx, symbol)
}

func (a *Adapter) storeOrderContext(orderID string, meta orderContext) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return
	}
	meta.OrderID = orderID
	if meta.UpdatedAt.IsZero() {
		meta.UpdatedAt = time.Now()
	}
	a.mu.Lock()
	a.orders[orderID] = meta
	a.compactOrderContextsLocked(maxPersistedOrderContexts)
	a.mu.Unlock()
	_ = a.persistOrderContexts()
}

func (a *Adapter) getOrderContext(orderID string) (orderContext, bool) {
	a.mu.RLock()
	meta, ok := a.orders[strings.TrimSpace(orderID)]
	a.mu.RUnlock()
	return meta, ok
}

func (a *Adapter) resolveOrderContext(ctx context.Context, orderID string) (orderContext, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return orderContext{}, broker.ErrInvalidOrderRequest
	}
	if meta, ok := a.getOrderContext(orderID); ok {
		return meta, nil
	}

	rows, err := a.fetchUnsettled(ctx, "")
	if err != nil {
		return orderContext{}, err
	}
	for _, row := range rows {
		if strings.TrimSpace(row.OrderNumber) != orderID {
			continue
		}
		meta := orderContext{
			OrderID:      orderID,
			Symbol:       normalizeSymbol(row.StockCode),
			Exchange:     exchangeFromUnsettledRow(row),
			Quantity:     row.OrderQty,
			RemainingQty: row.UnsettledQty,
			Price:        normalizedPrice(row.OrderPrice),
			Side:         sideFromKiwoomOrderText(row.OrderSideText),
			Status:       mapOrderStatus(row.OrderStatus, row.UnsettledQty),
			UpdatedAt:    time.Now(),
		}
		a.storeOrderContext(orderID, meta)
		return meta, nil
	}
	return orderContext{}, broker.ErrOrderNotFound
}

func toKiwoomExchange(market string) (string, error) {
	m := strings.ToUpper(strings.TrimSpace(market))
	if m == "" {
		return "KRX", nil
	}
	switch m {
	case "KR", "KRX", "KOSPI", "KOSDAQ":
		return "KRX", nil
	case "NXT":
		return "NXT", nil
	case "SOR":
		return "SOR", nil
	default:
		return "", broker.ErrInvalidMarket
	}
}

func toTradeTypeAndPrice(orderType broker.OrderType, price float64) (string, string, error) {
	switch orderType {
	case broker.OrderTypeMarket:
		return "3", "", nil
	case broker.OrderTypeLimit:
		if price <= 0 {
			return "", "", broker.ErrInvalidOrderRequest
		}
		return "0", formatPrice(price), nil
	default:
		return "", "", broker.ErrInvalidOrderRequest
	}
}

func formatPrice(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func normalizeSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	s = strings.TrimPrefix(s, "A")
	return s
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func mapOrderStatus(raw string, remaining int64) broker.OrderStatus {
	r := strings.TrimSpace(raw)
	switch {
	case strings.Contains(r, "취소"):
		return broker.OrderStatusCancelled
	case strings.Contains(r, "거부") || strings.Contains(r, "실패"):
		return broker.OrderStatusRejected
	case strings.Contains(r, "체결") && remaining == 0:
		return broker.OrderStatusFilled
	default:
		return broker.OrderStatusPending
	}
}

func exchangeFromUnsettledRow(row kiwoom.UnsettledOrder) string {
	if txt := strings.ToUpper(strings.TrimSpace(row.ExchangeText)); txt != "" {
		switch txt {
		case "KRX", "NXT", "SOR":
			return txt
		}
	}
	return mapStexCodeToMarket(row.ExchangeCode, row.ExchangeText)
}

func mapStexCodeToMarket(code, text string) string {
	switch strings.TrimSpace(code) {
	case "1":
		return "KRX"
	case "2":
		return "NXT"
	case "0":
		if strings.EqualFold(strings.TrimSpace(text), "SOR") {
			return "SOR"
		}
		return "KRX"
	}
	t := strings.ToUpper(strings.TrimSpace(text))
	if t == "" {
		return "KRX"
	}
	return t
}

func sideFromKiwoomText(v string) string {
	raw := strings.TrimSpace(v)
	switch {
	case strings.Contains(raw, "매수"):
		return string(broker.OrderSideBuy)
	case strings.Contains(raw, "매도"):
		return string(broker.OrderSideSell)
	default:
		return ""
	}
}

func sideFromKiwoomOrderText(v string) broker.OrderSide {
	if strings.Contains(strings.TrimSpace(v), "매도") {
		return broker.OrderSideSell
	}
	return broker.OrderSideBuy
}

func parseOrderTime(v string) time.Time {
	v = strings.TrimSpace(v)
	if len(v) == 6 {
		now := time.Now()
		parsed, err := time.ParseInLocation("150405", v, now.Location())
		if err == nil {
			return time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, now.Location())
		}
	}
	return time.Now()
}

func normalizeOutputMarket(market string) string {
	m := strings.ToUpper(strings.TrimSpace(market))
	switch m {
	case "KR", "KOSPI", "KOSDAQ", "":
		return "KRX"
	case "NXT", "SOR", "KRX":
		return m
	default:
		return m
	}
}

func marketFromCode(code, name string) string {
	switch strings.TrimSpace(code) {
	case "0", "10":
		return "KRX"
	case "30":
		return "KONEX"
	}
	if strings.Contains(strings.TrimSpace(name), "코스닥") {
		return "KRX"
	}
	return "KRX"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizedPrice(v float64) float64 {
	return math.Abs(v)
}
