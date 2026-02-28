package adapter

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/smallfish06/krsec/internal/kis"
	"github.com/smallfish06/krsec/pkg/broker"
)

// Adapter adapts KIS raw API to broker.Broker interface
type Adapter struct {
	client        *kis.Client
	accountID     string
	accountPrdtCD string // 계좌 상품 코드 (예: "01")
	sandbox       bool
	orderDir      string

	mu     sync.RWMutex
	orders map[string]orderContext // key: order id
}

// Options configures adapter internals such as token and persistence strategy.
type Options struct {
	TokenManager    kis.TokenManager
	OrderContextDir string
}

type orderContext struct {
	CANO         string
	AccountPrdt  string
	OrderID      string
	OrderOrgNo   string
	OrderDvsn    string
	OrderQty     int
	OrderPrice   float64
	ExchangeCode string
	IsOverseas   bool
	Symbol       string
	Status       broker.OrderStatus
	UpdatedAt    time.Time
}

// NewAdapterWithOptions creates a new KIS adapter with injectable dependencies.
func NewAdapterWithOptions(sandbox bool, accountID string, opts Options) *Adapter {
	// accountID 형식: "12345678-01" 또는 "12345678"
	// 분리: CANO = "12345678", ACNT_PRDT_CD = "01"
	cano := accountID
	acntPrdtCD := "01"

	// "-"로 분리되어 있으면 분리
	if len(accountID) > 2 && accountID[len(accountID)-3] == '-' {
		cano = accountID[:len(accountID)-3]
		acntPrdtCD = accountID[len(accountID)-2:]
	}

	a := &Adapter{
		client:        kis.NewClientWithTokenManager(sandbox, opts.TokenManager),
		accountID:     cano,
		accountPrdtCD: acntPrdtCD,
		sandbox:       sandbox,
		orderDir:      strings.TrimSpace(opts.OrderContextDir),
		orders:        make(map[string]orderContext),
	}
	if err := a.loadOrderContexts(); err != nil {
		log.Printf("Warning: failed to load persisted orders for %s-%s: %v", cano, acntPrdtCD, err)
	}
	return a
}

// Name returns the broker name
func (a *Adapter) Name() string {
	return "KIS"
}

// Authenticate authenticates with the broker
func (a *Adapter) Authenticate(ctx context.Context, creds broker.Credentials) (*broker.Token, error) {
	return a.client.Authenticate(ctx, creds)
}

// GetQuote retrieves a quote for a given market and symbol
// For overseas markets (us, us-nyse, us-nasdaq, us-amex), uses InquireOverseasPrice.
func (a *Adapter) GetQuote(ctx context.Context, market, symbol string) (*broker.Quote, error) {
	if quoteExcg, ok := toKISOverseasQuoteExchange(market); ok {
		return a.getOverseasQuote(ctx, market, symbol, quoteExcg)
	}

	resp, err := a.client.InquirePrice(ctx, market, symbol)
	if err != nil {
		return nil, err
	}

	// 응답 데이터가 output 또는 output1에 있을 수 있음
	output := resp.Output
	if output.StckPrpr == "" && resp.Output1.StckPrpr != "" {
		output = resp.Output1
	}

	price, _ := strconv.ParseFloat(output.StckPrpr, 64)
	open, _ := strconv.ParseFloat(output.StckOprc, 64)
	high, _ := strconv.ParseFloat(output.StckHgpr, 64)
	low, _ := strconv.ParseFloat(output.StckLwpr, 64)
	change := parseFirstFloat(output.PrdyVrss)
	prevClose := 0.0
	if price != 0 || change != 0 {
		prevClose = price - change
	}
	volume, _ := strconv.ParseInt(output.AcmlVol, 10, 64)

	return &broker.Quote{
		Symbol:     symbol,
		Market:     market,
		Price:      price,
		Open:       open,
		High:       high,
		Low:        low,
		Close:      price, // 현재가를 종가로 사용
		PrevClose:  prevClose,
		Change:     change,
		ChangeRate: parseFirstFloat(output.PrdyCtrt),
		Volume:     volume,
		Turnover:   parseFirstFloat(output.AcmlTrPbmn),
		UpperLimit: parseFirstFloat(output.StckMxpr),
		LowerLimit: parseFirstFloat(output.StckLlam),
		Timestamp:  time.Now(),
	}, nil
}

// getOverseasQuote retrieves overseas stock quote via KIS API (HHDFS00000300)
func (a *Adapter) getOverseasQuote(ctx context.Context, market, symbol, exchangeCode string) (*broker.Quote, error) {
	resp, err := a.client.InquireOverseasPrice(ctx, exchangeCode, symbol)
	if err != nil {
		return nil, err
	}

	price, _ := strconv.ParseFloat(resp.Output.Last, 64)
	open, _ := strconv.ParseFloat(resp.Output.Open, 64)
	high, _ := strconv.ParseFloat(resp.Output.High, 64)
	low, _ := strconv.ParseFloat(resp.Output.Low, 64)
	change := parseFirstFloat(resp.Output.PrdyVrss)
	prevClose := 0.0
	if price != 0 || change != 0 {
		prevClose = price - change
	}
	volume, _ := strconv.ParseInt(resp.Output.AccrTrVol, 10, 64)

	return &broker.Quote{
		Symbol:    symbol,
		Market:    market,
		Price:     price,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     price,
		PrevClose: prevClose,
		Change:    change,
		Volume:    volume,
		Timestamp: time.Now(),
	}, nil
}

// GetOHLCV retrieves OHLCV data for a given market and symbol
func (a *Adapter) GetOHLCV(ctx context.Context, market, symbol string, opts broker.OHLCVOpts) ([]broker.OHLCV, error) {
	resp, err := a.client.InquireDailyPrice(ctx, market, symbol, "", "", true)
	if err != nil {
		return nil, err
	}

	result := make([]broker.OHLCV, 0, len(resp.Output))
	for _, item := range resp.Output {
		timestamp, _ := time.Parse("20060102", item.StckBsopDate)
		open, _ := strconv.ParseFloat(item.StckOprc, 64)
		high, _ := strconv.ParseFloat(item.StckHgpr, 64)
		low, _ := strconv.ParseFloat(item.StckLwpr, 64)
		close, _ := strconv.ParseFloat(item.StckClpr, 64)
		volume, _ := strconv.ParseInt(item.AcmlVol, 10, 64)

		result = append(result, broker.OHLCV{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return applyOHLCVOptions(result, opts)
}

// GetBalance retrieves account balance
func (a *Adapter) GetBalance(ctx context.Context, accountID string) (*broker.Balance, error) {
	// accountID 파싱
	cano, acntPrdtCD := a.parseAccountID(accountID)

	resp, err := a.client.InquireBalance(ctx, cano, acntPrdtCD)
	if err != nil {
		return nil, err
	}

	if len(resp.Output2) == 0 {
		return &broker.Balance{AccountID: accountID}, nil
	}
	summary := resp.Output2[0]

	return toBrokerBalance(accountID, summary), nil
}

func toBrokerBalance(accountID string, summary kis.StockBalanceSummary) *broker.Balance {
	cash := parseFirstFloat(summary.DnpspCblAmt)
	return &broker.Balance{
		AccountID:        accountID,
		Cash:             cash,
		TotalAssets:      parseFirstFloat(summary.DpsplTotAmt),
		BuyingPower:      cash, // KIS 잔고 응답에는 별도 주문가능현금 필드가 없어 예수금을 사용
		WithdrawableCash: cash,
		ReceivableAmount: parseFirstFloat(summary.PrvsRcdlExcc),
		ProfitLoss:       parseFirstFloat(summary.TotEvluPfls, summary.EvluPflsSm),
		ProfitLossPct:    parseFirstFloat(summary.EvluErngRt),
		PositionCost:     parseFirstFloat(summary.PchsAmtSmtl),
		PositionValue:    parseFirstFloat(summary.EvluAmtSmtl),
		SettlementT1:     parseFirstFloat(summary.NxdyExccAmt),
		Unsettled:        parseFirstFloat(summary.PrvsRcdlExcc),
		LoanBalance:      parseFirstFloat(summary.TotStlnSlCa),
	}
}

func toBrokerStockPosition(item kis.StockBalanceOutput) broker.Position {
	qty, _ := strconv.ParseInt(item.HldgQty, 10, 64)
	orderableQty, _ := strconv.ParseInt(item.OrdPsblQty, 10, 64)
	return broker.Position{
		Symbol:        item.PdNo,
		Name:          item.PrdtName,
		Market:        "KRX",
		AssetType:     broker.AssetStock,
		Quantity:      qty,
		OrderableQty:  orderableQty,
		AvgPrice:      parseFirstFloat(item.PchmAvgPric),
		CurrentPrice:  parseFirstFloat(item.Prpr, item.PrprTprt),
		PurchaseValue: parseFirstFloat(item.PchmAmt),
		MarketValue:   parseFirstFloat(item.EvluAmt),
		ProfitLoss:    parseFirstFloat(item.EvluPflsAmt),
		ProfitLossPct: parseFirstFloat(item.EvluPflsRt),
		WeightPct:     parseFirstFloat(item.HldgQtyRatio),
	}
}

// GetPositions retrieves account positions (stocks + bonds)
func (a *Adapter) GetPositions(ctx context.Context, accountID string) ([]broker.Position, error) {
	cano, acntPrdtCD := a.parseAccountID(accountID)

	var positions []broker.Position

	// 1. 주식 잔고 조회
	resp, err := a.client.InquireBalance(ctx, cano, acntPrdtCD)
	if err == nil {
		for _, item := range resp.Output1 {
			positions = append(positions, toBrokerStockPosition(item))
		}
	}

	// 2. 채권 잔고 조회 (실패해도 주식 결과는 반환)
	bondResp, err := a.client.InquireBondBalance(ctx, cano, acntPrdtCD)
	if err != nil {
		log.Printf("bond balance query failed (non-fatal): %v", err)
	}
	if err == nil {
		bondItems := bondResp.Output
		if len(bondItems) == 0 {
			bondItems = bondResp.Output1
		}
		// Aggregate by symbol (same bond bought on different dates)
		type bondAgg struct {
			name     string
			totalQty int64
			totalAmt float64
		}
		bondMap := make(map[string]*bondAgg)
		for _, item := range bondItems {
			qty, _ := strconv.ParseInt(item.CblcQty, 10, 64)
			if qty == 0 {
				continue
			}
			buyAmt, _ := strconv.ParseFloat(item.BuyAmt, 64)
			if agg, ok := bondMap[item.PdNo]; ok {
				agg.totalQty += qty
				agg.totalAmt += buyAmt
			} else {
				bondMap[item.PdNo] = &bondAgg{name: item.PrdtName, totalQty: qty, totalAmt: buyAmt}
			}
		}
		for symbol, agg := range bondMap {
			avgPrice := agg.totalAmt / float64(agg.totalQty)
			positions = append(positions, broker.Position{
				Symbol:        symbol,
				Name:          agg.name,
				Market:        "KRX",
				AssetType:     broker.AssetBond,
				Quantity:      agg.totalQty,
				AvgPrice:      avgPrice,
				CurrentPrice:  avgPrice,
				PurchaseValue: agg.totalAmt,
				ProfitLoss:    0,
			})
		}
	}

	if positions == nil {
		positions = []broker.Position{}
	}

	return positions, nil
}

// GetInstrument retrieves normalized instrument metadata.
func (a *Adapter) GetInstrument(ctx context.Context, market, symbol string) (*broker.Instrument, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}

	if cached, ok := kis.LookupMasterSymbol(market, symbol); ok {
		return &broker.Instrument{
			Symbol:          cached.Symbol,
			Market:          cached.Market,
			Name:            firstNonEmpty(cached.Name, symbol),
			NameEn:          cached.NameEn,
			Exchange:        cached.Exchange,
			Currency:        cached.Currency,
			Country:         cached.Country,
			AssetType:       toBrokerAssetType(cached.Market, cached.ProductType),
			ProductType:     cached.ProductType,
			ProductTypeCode: cached.ProductTypeCode,
			SecurityGroup:   cached.SecurityGroup,
			IsListed:        cached.IsListed,
			IsSuspended:     false,
		}, nil
	}

	prdtTypeCode, isOverseas, err := toKISProductTypeCode(market)
	if err != nil {
		return nil, err
	}

	if isOverseas {
		resp, err := a.client.InquireOverseasProductBasicInfo(ctx, symbol, prdtTypeCode)
		if err != nil {
			return nil, err
		}
		out := resp.Output
		if out.StdPdNo == "" && out.PrdtName == "" && out.PrdtEngName == "" {
			return nil, broker.ErrInstrumentNotFound
		}

		name := out.PrdtName
		if name == "" {
			name = out.OvrsItemName
		}
		if name == "" {
			name = symbol
		}

		isListed := parseYN(out.LstgYN)
		if parseYN(out.LstgAbolItemYN) {
			isListed = false
		}

		return &broker.Instrument{
			Symbol:          symbol,
			Market:          strings.ToUpper(strings.TrimSpace(market)),
			ISIN:            normalizeISIN(out.StdPdNo),
			Name:            name,
			NameEn:          out.PrdtEngName,
			Exchange:        out.OvrsExcgCD,
			Currency:        out.TrCrcyCD,
			Country:         out.NatnName,
			AssetType:       broker.AssetOverseas,
			ProductType:     overseasProductTypeFromCode(out.OvrsStckDvsnCD),
			ProductTypeCode: prdtTypeCode,
			SecurityGroup:   out.OvrsStckDvsnCD,
			Sector:          out.PrdtClsfName,
			IsListed:        isListed,
			IsSuspended:     out.OvrsStckTrStopDvsnCD != "" && out.OvrsStckTrStopDvsnCD != "01",
			ListingDate:     out.LstgDt,
			DelistingDate:   out.LstgAbolDt,
		}, nil
	}

	resp, err := a.client.InquireStockBasicInfo(ctx, symbol, prdtTypeCode)
	if err == nil {
		out := resp.Output
		if out.PdNo == "" && out.PrdtName == "" {
			return nil, broker.ErrInstrumentNotFound
		}

		return &broker.Instrument{
			Symbol:          firstNonEmpty(out.PdNo, symbol),
			Market:          normalizeDomesticMarket(strings.ToUpper(strings.TrimSpace(market)), out.MketIDCD),
			ISIN:            normalizeISIN(out.StdPdNo),
			Name:            firstNonEmpty(out.PrdtName, symbol),
			NameEn:          out.PrdtEngName,
			ShortName:       out.PrdtAbrvName,
			Exchange:        normalizeDomesticExchange(out.ExcgDvsnCD),
			Currency:        "KRW",
			Country:         "KR",
			AssetType:       domesticAssetTypeFromSecurityGroup(out.SctyGrpIDCD),
			ProductType:     domesticProductTypeFromSecurityGroup(out.SctyGrpIDCD),
			ProductTypeCode: out.PrdtTypeCD,
			SecurityGroup:   out.SctyGrpIDCD,
			Sector:          firstNonEmpty(out.StdIdstClsfCDName, out.PrdtClsfName),
			ListedShares:    int64(parseIntOrDefault(out.LstgStqt, 0)),
			IsListed:        out.LstgAbolDt == "",
			IsSuspended:     parseYN(out.TrStopYN),
			ListingDate:     firstNonEmpty(out.SctsMketLstgDt, out.KosdaqMketLstgDt, out.FrbdMketLstgDt),
			DelistingDate:   out.LstgAbolDt,
		}, nil
	}

	// Fallback to the more generic domestic product info API when stock-info is unavailable.
	fallback, fallbackErr := a.client.InquireProductBasicInfo(ctx, symbol, prdtTypeCode)
	if fallbackErr != nil {
		return nil, err
	}
	out := fallback.Output
	if out.PdNo == "" && out.PrdtName == "" {
		return nil, broker.ErrInstrumentNotFound
	}

	return &broker.Instrument{
		Symbol:          firstNonEmpty(out.PdNo, symbol),
		Market:          strings.ToUpper(strings.TrimSpace(market)),
		ISIN:            normalizeISIN(out.StdPdNo),
		Name:            firstNonEmpty(out.PrdtName, symbol),
		NameEn:          out.PrdtEngName,
		ShortName:       out.PrdtAbrvName,
		Currency:        "KRW",
		Country:         "KR",
		AssetType:       broker.AssetStock,
		ProductTypeCode: out.PrdtTypeCD,
		Sector:          out.PrdtClsfName,
		IsListed:        out.SaleEndDt == "",
		IsSuspended:     false,
		ListingDate:     out.SaleStrtDt,
		DelistingDate:   out.SaleEndDt,
	}, nil
}

// BootstrapSymbols loads KIS master symbol files into memory for fast lookups.
func (a *Adapter) BootstrapSymbols(ctx context.Context) (int, error) {
	return a.client.BootstrapMasterSymbols(ctx)
}

// ReloadSymbols force-reloads KIS master symbol files.
func (a *Adapter) ReloadSymbols(ctx context.Context) (int, error) {
	return a.client.ReloadMasterSymbols(ctx)
}

// PlaceOrder places a new order
func (a *Adapter) PlaceOrder(ctx context.Context, req broker.OrderRequest) (*broker.OrderResult, error) {
	orderType := "limit"
	orderDvsn := "00"
	if req.Type == broker.OrderTypeMarket {
		orderType = "market"
		orderDvsn = "01"
	}

	side := "buy"
	if req.Side == broker.OrderSideSell {
		side = "sell"
	}

	cano, acntPrdtCD := a.parseAccountID(req.AccountID)
	if ovrsExcg, ok := toKISOverseasExchange(req.Market); ok {
		if req.Type != broker.OrderTypeLimit {
			return nil, broker.ErrInvalidOrderRequest
		}

		resp, err := a.client.OrderOverseas(ctx, cano, acntPrdtCD, ovrsExcg, req.Symbol, int(req.Quantity), req.Price, side, "00")
		if err != nil {
			return nil, err
		}

		a.storeOrderContext(resp.Output.OrdNo, orderContext{
			CANO:         cano,
			AccountPrdt:  acntPrdtCD,
			OrderID:      resp.Output.OrdNo,
			OrderDvsn:    "00",
			OrderQty:     int(req.Quantity),
			OrderPrice:   req.Price,
			ExchangeCode: ovrsExcg,
			IsOverseas:   true,
			Symbol:       req.Symbol,
			Status:       broker.OrderStatusPending,
			UpdatedAt:    time.Now(),
		})

		return &broker.OrderResult{
			OrderID:        resp.Output.OrdNo,
			Status:         broker.OrderStatusPending,
			RemainingQty:   req.Quantity,
			AvgFilledPrice: 0,
			Message:        resp.Msg1,
			Timestamp:      time.Now(),
		}, nil
	}

	exchangeCode := toKISExchangeID(req.Market)

	resp, err := a.client.OrderCash(ctx, cano, acntPrdtCD, req.Symbol, orderType, int(req.Quantity), int(req.Price), side, exchangeCode)
	if err != nil {
		return nil, err
	}

	a.storeOrderContext(resp.Output.OrdNo, orderContext{
		CANO:         cano,
		AccountPrdt:  acntPrdtCD,
		OrderID:      resp.Output.OrdNo,
		OrderOrgNo:   resp.Output.KrxFwdOrdOrgno,
		OrderDvsn:    orderDvsn,
		OrderQty:     int(req.Quantity),
		OrderPrice:   req.Price,
		ExchangeCode: exchangeCode,
		Symbol:       req.Symbol,
		Status:       broker.OrderStatusPending,
		UpdatedAt:    time.Now(),
	})

	return &broker.OrderResult{
		OrderID:        resp.Output.OrdNo,
		Status:         broker.OrderStatusPending,
		RemainingQty:   req.Quantity,
		AvgFilledPrice: 0,
		Message:        resp.Msg1,
		Timestamp:      time.Now(),
	}, nil
}

// CancelOrder cancels an order
func (a *Adapter) CancelOrder(ctx context.Context, orderID string) error {
	meta, err := a.resolveOrderContext(ctx, orderID)
	if err != nil {
		return err
	}

	if meta.IsOverseas {
		_, err = a.client.OrderOverseasRvseCncl(
			ctx,
			meta.CANO,
			meta.AccountPrdt,
			meta.ExchangeCode,
			meta.Symbol,
			meta.OrderID,
			"02",
			meta.OrderQty,
			0,
		)
	} else {
		_, err = a.client.OrderRvseCncl(
			ctx,
			meta.CANO,
			meta.AccountPrdt,
			meta.OrderOrgNo,
			meta.OrderID,
			meta.OrderDvsn,
			"02",
			meta.OrderQty,
			int(meta.OrderPrice),
			true,
			meta.ExchangeCode,
		)
	}
	if err != nil {
		return err
	}

	meta.Status = broker.OrderStatusCancelled
	meta.UpdatedAt = time.Now()
	a.storeOrderContext(meta.OrderID, meta)
	if meta.OrderID != orderID {
		a.storeOrderContext(orderID, meta)
	}
	return nil
}

// ModifyOrder modifies an existing order
func (a *Adapter) ModifyOrder(ctx context.Context, orderID string, req broker.ModifyOrderRequest) (*broker.OrderResult, error) {
	meta, err := a.resolveOrderContext(ctx, orderID)
	if err != nil {
		return nil, err
	}

	newQty := meta.OrderQty
	if req.Quantity > 0 {
		newQty = int(req.Quantity)
	}
	if newQty <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}

	newPrice := meta.OrderPrice
	if req.Price > 0 {
		newPrice = req.Price
	}

	var resp *kis.OrderResponse
	if meta.IsOverseas {
		if newPrice <= 0 {
			return nil, broker.ErrInvalidOrderRequest
		}
		resp, err = a.client.OrderOverseasRvseCncl(
			ctx,
			meta.CANO,
			meta.AccountPrdt,
			meta.ExchangeCode,
			meta.Symbol,
			meta.OrderID,
			"01",
			newQty,
			newPrice,
		)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err = a.client.OrderRvseCncl(
			ctx,
			meta.CANO,
			meta.AccountPrdt,
			meta.OrderOrgNo,
			meta.OrderID,
			meta.OrderDvsn,
			"01",
			newQty,
			int(newPrice),
			false,
			meta.ExchangeCode,
		)
		if err != nil {
			return nil, err
		}
	}

	newOrderID := resp.Output.OrdNo
	if newOrderID == "" {
		newOrderID = orderID
	}

	a.storeOrderContext(newOrderID, orderContext{
		CANO:         meta.CANO,
		AccountPrdt:  meta.AccountPrdt,
		OrderID:      newOrderID,
		OrderOrgNo:   resp.Output.KrxFwdOrdOrgno,
		OrderDvsn:    meta.OrderDvsn,
		OrderQty:     newQty,
		OrderPrice:   newPrice,
		ExchangeCode: meta.ExchangeCode,
		IsOverseas:   meta.IsOverseas,
		Symbol:       meta.Symbol,
		Status:       broker.OrderStatusPending,
		UpdatedAt:    time.Now(),
	})
	if newOrderID != orderID {
		a.removeOrderContext(orderID)
	}

	return &broker.OrderResult{
		OrderID:      newOrderID,
		Status:       broker.OrderStatusPending,
		RemainingQty: int64(newQty),
		Message:      resp.Msg1,
		Timestamp:    time.Now(),
	}, nil
}

// GetOrder returns current order status based on cached/meta information.
func (a *Adapter) GetOrder(ctx context.Context, orderID string) (*broker.OrderResult, error) {
	meta, err := a.resolveOrderContext(ctx, orderID)
	if err != nil {
		return nil, err
	}

	status := meta.Status
	if status == "" {
		status = broker.OrderStatusPending
	}
	var filledQty int64
	var remainQty int64
	var avgFilledPrice float64
	var rejectReason string

	if status != broker.OrderStatusCancelled {
		snap, ok := a.resolveRemoteOrderSnapshot(ctx, meta)
		if ok {
			status = snap.Status
			filledQty = snap.FilledQty
			remainQty = snap.RemainingQty
			avgFilledPrice = snap.AvgFilledPrice
			rejectReason = snap.RejectedReason
		} else if !meta.IsOverseas {
			// Fallback: check current open-order list only when daily fill lookup is unavailable.
			resp, err := a.client.InquirePossibleRvseCncl(ctx, meta.CANO, meta.AccountPrdt)
			if err == nil {
				found := false
				for _, item := range resp.Output {
					if item.ODNo == meta.OrderID || item.OrgnODNo == meta.OrderID {
						found = true
						break
					}
				}
				if found {
					status = broker.OrderStatusPending
				} else if status == broker.OrderStatusPending {
					status = broker.OrderStatusFilled
				}
			}
		}
	}
	if remainQty == 0 {
		switch status {
		case broker.OrderStatusPending:
			if meta.OrderQty > 0 {
				remainQty = int64(meta.OrderQty)
			}
		case broker.OrderStatusFilled, broker.OrderStatusCancelled:
			remainQty = 0
		}
	}
	if filledQty == 0 && status == broker.OrderStatusFilled && meta.OrderQty > 0 {
		filledQty = int64(meta.OrderQty)
	}

	timestamp := meta.UpdatedAt
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	if status != meta.Status {
		meta.Status = status
		meta.UpdatedAt = time.Now()
		timestamp = meta.UpdatedAt
		a.storeOrderContext(meta.OrderID, meta)
	}

	return &broker.OrderResult{
		OrderID:        meta.OrderID,
		Status:         status,
		FilledQuantity: filledQty,
		RemainingQty:   remainQty,
		AvgFilledPrice: avgFilledPrice,
		RejectedReason: rejectReason,
		Timestamp:      timestamp,
	}, nil
}

// GetOrderFills returns normalized fill executions for an order.
func (a *Adapter) GetOrderFills(ctx context.Context, orderID string) ([]broker.OrderFill, error) {
	meta, err := a.resolveOrderContext(ctx, orderID)
	if err != nil {
		return nil, err
	}

	startDate := time.Now().AddDate(0, 0, -7)
	if !meta.UpdatedAt.IsZero() {
		startDate = meta.UpdatedAt.AddDate(0, 0, -2)
	}
	if startDate.After(time.Now()) {
		startDate = time.Now().AddDate(0, 0, -1)
	}
	start := startDate.Format("20060102")
	end := time.Now().Format("20060102")

	if meta.IsOverseas {
		exchange := meta.ExchangeCode
		if exchange == "" {
			exchange = "%"
		}
		resp, err := a.client.InquireOverseasCcnl(ctx, meta.CANO, meta.AccountPrdt, start, end, exchange)
		if err != nil {
			return nil, err
		}
		fills := make([]broker.OrderFill, 0)
		for _, item := range resp.Output {
			if item.ODNo != meta.OrderID && item.OrgnODNo != meta.OrderID {
				continue
			}
			qty := int64(parseIntOrDefault(item.FtCcldQty, 0))
			if qty <= 0 {
				continue
			}
			price, _ := strconv.ParseFloat(strings.TrimSpace(item.FtCcldUNPR3), 64)
			amount, _ := strconv.ParseFloat(strings.TrimSpace(item.FtCcldAmt3), 64)
			if amount == 0 && price > 0 {
				amount = float64(qty) * price
			}
			filledAt := parseOrderDateTime(item.OrdDt, item.OrdTmd)
			fills = append(fills, broker.OrderFill{
				OrderID:   meta.OrderID,
				Symbol:    firstNonEmpty(item.PdNo, meta.Symbol),
				Market:    strings.ToUpper(strings.TrimSpace(meta.ExchangeCode)),
				Side:      normalizeSideLabel(item.SllBuyDvsnCD),
				Quantity:  qty,
				Price:     price,
				Amount:    amount,
				Currency:  item.TrCrcyCD,
				FilledAt:  filledAt,
				RawStatus: item.PrcsStatName,
			})
		}
		if len(fills) == 0 {
			return []broker.OrderFill{}, nil
		}
		return fills, nil
	}

	exchangeID := meta.ExchangeCode
	if exchangeID == "" {
		exchangeID = "ALL"
	}
	resp, err := a.client.InquireDailyCcld(ctx, meta.CANO, meta.AccountPrdt, start, end, meta.OrderOrgNo, meta.OrderID, exchangeID)
	if err != nil {
		return nil, err
	}

	fills := make([]broker.OrderFill, 0)
	for _, item := range resp.Output1 {
		if item.ODNo != meta.OrderID && item.OrgnODNo != meta.OrderID {
			continue
		}
		qty := int64(parseIntOrDefault(item.TotCcldQty, 0))
		if qty <= 0 {
			continue
		}
		price, _ := strconv.ParseFloat(strings.TrimSpace(item.AvgPrvs), 64)
		if price == 0 {
			price, _ = strconv.ParseFloat(strings.TrimSpace(item.OrdUNPR), 64)
		}
		amount, _ := strconv.ParseFloat(strings.TrimSpace(item.TotCcldAmt), 64)
		if amount == 0 && price > 0 {
			amount = float64(qty) * price
		}
		filledAt := parseOrderDateTime(item.OrdDt, item.OrdTmd)

		fills = append(fills, broker.OrderFill{
			OrderID:   meta.OrderID,
			Symbol:    firstNonEmpty(item.PdNo, meta.Symbol),
			Market:    firstNonEmpty(item.ExcgIDDvsn, meta.ExchangeCode, "KRX"),
			Side:      normalizeSideLabel(item.SllBuyDvsn),
			Quantity:  qty,
			Price:     price,
			Amount:    amount,
			Currency:  "KRW",
			FilledAt:  filledAt,
			RawStatus: item.OrdDvsnCD,
		})
	}
	if fills == nil {
		fills = []broker.OrderFill{}
	}
	return fills, nil
}

// parseAccountID parses account ID into CANO and ACNT_PRDT_CD
func (a *Adapter) parseAccountID(accountID string) (string, string) {
	// 어댑터에 기본 계좌가 설정되어 있으면 그것을 사용
	if accountID == "" || accountID == a.accountID || accountID == a.accountID+"-"+a.accountPrdtCD {
		return a.accountID, a.accountPrdtCD
	}

	// accountID 형식: "12345678-01" 또는 "12345678"
	cano := accountID
	acntPrdtCD := "01"

	if len(accountID) > 2 && accountID[len(accountID)-3] == '-' {
		cano = accountID[:len(accountID)-3]
		acntPrdtCD = accountID[len(accountID)-2:]
	}

	return cano, acntPrdtCD
}

func (a *Adapter) storeOrderContext(orderID string, meta orderContext) {
	if orderID == "" {
		return
	}
	a.mu.Lock()
	a.orders[orderID] = meta
	a.compactOrderContextsLocked(maxPersistedOrderContexts)
	a.mu.Unlock()

	if err := a.persistOrderContexts(); err != nil {
		log.Printf("Warning: failed to persist order contexts for %s-%s: %v", a.accountID, a.accountPrdtCD, err)
	}
}

func (a *Adapter) getOrderContext(orderID string) (orderContext, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	meta, ok := a.orders[orderID]
	return meta, ok
}

func (a *Adapter) removeOrderContext(orderID string) {
	a.mu.Lock()
	delete(a.orders, orderID)
	a.mu.Unlock()

	if err := a.persistOrderContexts(); err != nil {
		log.Printf("Warning: failed to persist order contexts for %s-%s: %v", a.accountID, a.accountPrdtCD, err)
	}
}

func (a *Adapter) resolveOrderContext(ctx context.Context, orderID string) (orderContext, error) {
	if orderID == "" {
		return orderContext{}, broker.ErrInvalidOrderRequest
	}
	cached, hasCached := a.getOrderContext(orderID)
	if hasCached && (cached.IsOverseas || cached.OrderOrgNo != "") {
		return cached, nil
	}

	resp, err := a.client.InquirePossibleRvseCncl(ctx, a.accountID, a.accountPrdtCD)
	if err != nil {
		if hasCached {
			return cached, nil
		}
		return orderContext{}, err
	}
	for _, item := range resp.Output {
		if item.ODNo != orderID && item.OrgnODNo != orderID {
			continue
		}

		qty, _ := strconv.Atoi(item.PsblQty)
		if qty == 0 {
			qty, _ = strconv.Atoi(item.OrdQty)
		}
		price, _ := strconv.Atoi(item.OrdUNPR)
		meta := orderContext{
			CANO:         a.accountID,
			AccountPrdt:  a.accountPrdtCD,
			OrderID:      item.ODNo,
			OrderOrgNo:   item.OrdGnoBrno,
			OrderDvsn:    item.OrdDvsn,
			OrderQty:     qty,
			OrderPrice:   float64(price),
			ExchangeCode: item.ExcgIDDvsnCD,
			Status:       broker.OrderStatusPending,
			UpdatedAt:    time.Now(),
		}
		if hasCached {
			if meta.OrderQty == 0 {
				meta.OrderQty = cached.OrderQty
			}
			if meta.OrderPrice == 0 {
				meta.OrderPrice = cached.OrderPrice
			}
			if meta.OrderDvsn == "" {
				meta.OrderDvsn = cached.OrderDvsn
			}
			if meta.ExchangeCode == "" {
				meta.ExchangeCode = cached.ExchangeCode
			}
			if meta.OrderOrgNo == "" {
				meta.OrderOrgNo = cached.OrderOrgNo
			}
			if meta.Symbol == "" {
				meta.Symbol = cached.Symbol
			}
			if meta.Status == "" {
				meta.Status = cached.Status
			}
			if !meta.IsOverseas {
				meta.IsOverseas = cached.IsOverseas
			}
		}
		if meta.OrderID == "" {
			meta.OrderID = orderID
		}
		if meta.OrderDvsn == "" {
			meta.OrderDvsn = "00"
		}
		if meta.ExchangeCode == "" {
			meta.ExchangeCode = "KRX"
		}
		a.storeOrderContext(meta.OrderID, meta)
		return meta, nil
	}

	if hasCached {
		return cached, nil
	}

	return orderContext{}, broker.ErrOrderNotFound
}

func toKISExchangeID(market string) string {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "", "KRX", "KOSPI", "KOSDAQ":
		return "KRX"
	case "NXT":
		return "NXT"
	case "SOR":
		return "SOR"
	default:
		return "KRX"
	}
}

type orderStatusSnapshot struct {
	Status         broker.OrderStatus
	FilledQty      int64
	RemainingQty   int64
	AvgFilledPrice float64
	RejectedReason string
}

func (a *Adapter) resolveRemoteOrderSnapshot(ctx context.Context, meta orderContext) (orderStatusSnapshot, bool) {
	startDate := time.Now().AddDate(0, 0, -7)
	if !meta.UpdatedAt.IsZero() {
		startDate = meta.UpdatedAt.AddDate(0, 0, -2)
	}
	if startDate.After(time.Now()) {
		startDate = time.Now().AddDate(0, 0, -1)
	}
	endDate := time.Now()
	start := startDate.Format("20060102")
	end := endDate.Format("20060102")

	if meta.IsOverseas {
		exchange := meta.ExchangeCode
		if exchange == "" {
			exchange = "%"
		}
		resp, err := a.client.InquireOverseasCcnl(ctx, meta.CANO, meta.AccountPrdt, start, end, exchange)
		if err != nil {
			return orderStatusSnapshot{}, false
		}
		for _, item := range resp.Output {
			if item.ODNo != meta.OrderID && item.OrgnODNo != meta.OrderID {
				continue
			}
			ordQty := parseIntOrDefault(item.FtOrdQty, meta.OrderQty)
			filledQty := parseIntOrDefault(item.FtCcldQty, 0)
			remainQty := parseIntOrDefault(item.NccsQty, 0)
			if remainQty == 0 && ordQty > 0 && filledQty >= 0 && filledQty < ordQty {
				remainQty = ordQty - filledQty
			}
			avgPrice := parseFirstFloat(item.FtCcldUNPR3, item.FtOrdUNPR3)
			base := orderStatusSnapshot{
				FilledQty:      int64(filledQty),
				RemainingQty:   int64(remainQty),
				AvgFilledPrice: avgPrice,
			}
			if strings.Contains(item.PrcsStatName, "거부") || strings.TrimSpace(item.RjctRson) != "" {
				base.Status = broker.OrderStatusRejected
				base.RejectedReason = firstNonEmpty(item.RjctRsonName, item.RjctRson, item.PrcsStatName)
				return base, true
			}
			if remainQty == 0 && filledQty > 0 && (ordQty == 0 || filledQty >= ordQty) {
				base.Status = broker.OrderStatusFilled
				return base, true
			}
			if meta.Status == broker.OrderStatusCancelled {
				base.Status = broker.OrderStatusCancelled
				return base, true
			}
			base.Status = broker.OrderStatusPending
			return base, true
		}
		return orderStatusSnapshot{}, false
	}

	exchangeID := meta.ExchangeCode
	if exchangeID == "" {
		exchangeID = "ALL"
	}
	resp, err := a.client.InquireDailyCcld(ctx, meta.CANO, meta.AccountPrdt, start, end, meta.OrderOrgNo, meta.OrderID, exchangeID)
	if err != nil {
		return orderStatusSnapshot{}, false
	}
	for _, item := range resp.Output1 {
		if item.ODNo != meta.OrderID && item.OrgnODNo != meta.OrderID {
			continue
		}

		ordQty := parseIntOrDefault(item.OrdQty, meta.OrderQty)
		filledQty := parseIntOrDefault(item.TotCcldQty, 0)
		remainQty := parseIntOrDefault(item.RmnQty, 0)
		if remainQty == 0 && ordQty > 0 && filledQty >= 0 && filledQty < ordQty {
			remainQty = ordQty - filledQty
		}
		rejectQty := parseIntOrDefault(item.RjctQty, 0)
		cancelled := parseYN(item.CnclYN) || parseIntOrDefault(item.CncCfrmQty, 0) > 0
		base := orderStatusSnapshot{
			FilledQty:      int64(filledQty),
			RemainingQty:   int64(remainQty),
			AvgFilledPrice: parseFirstFloat(item.AvgPrvs, item.OrdUNPR),
		}

		if rejectQty > 0 && filledQty == 0 {
			base.Status = broker.OrderStatusRejected
			base.RejectedReason = "rejected"
			return base, true
		}
		if cancelled {
			base.Status = broker.OrderStatusCancelled
			return base, true
		}
		if filledQty > 0 && (remainQty == 0 || (ordQty > 0 && filledQty >= ordQty)) {
			base.Status = broker.OrderStatusFilled
			return base, true
		}
		base.Status = broker.OrderStatusPending
		return base, true
	}

	return orderStatusSnapshot{}, false
}

func parseIntOrDefault(v string, d int) int {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return d
	}
	return n
}

func parseFirstFloat(v ...string) float64 {
	for _, raw := range v {
		s := strings.TrimSpace(raw)
		if s == "" {
			continue
		}
		n, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return n
		}
	}
	return 0
}

func parseOrderDateTime(date, t string) time.Time {
	d := strings.TrimSpace(date)
	if d == "" {
		return time.Time{}
	}
	ts := strings.TrimSpace(t)
	if len(ts) >= 6 {
		if v, err := time.Parse("20060102150405", d+ts[:6]); err == nil {
			return v
		}
	}
	if v, err := time.Parse("20060102", d); err == nil {
		return v
	}
	return time.Time{}
}

func normalizeSideLabel(code string) string {
	switch strings.TrimSpace(code) {
	case "01":
		return "sell"
	case "02":
		return "buy"
	default:
		return ""
	}
}

func toKISOverseasExchange(market string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "US", "US-NASDAQ", "NASDAQ", "NAS", "NASD":
		return "NASD", true
	case "US-NYSE", "NYSE", "NYS":
		return "NYSE", true
	case "US-AMEX", "AMEX", "AMS":
		return "AMEX", true
	case "HK", "HKEX", "HONGKONG", "SEHK", "HKS":
		return "SEHK", true
	case "JP", "JAPAN", "TSE", "JPX", "TKSE":
		return "TKSE", true
	case "SH", "SHA", "SHAA", "SHS", "SSE", "SHANGHAI":
		return "SHAA", true
	case "SZ", "SZA", "SZAA", "SZS", "SZSE", "SHENZHEN":
		return "SZAA", true
	case "HNX", "HASE", "HANOI":
		return "HASE", true
	case "HSX", "VNSE", "HOCHIMINH":
		return "VNSE", true
	default:
		return "", false
	}
}

func toKISOverseasQuoteExchange(market string) (string, bool) {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "US", "US-NASDAQ", "NASDAQ", "NAS", "NASD":
		return "NAS", true
	case "US-NYSE", "NYSE", "NYS":
		return "NYS", true
	case "US-AMEX", "AMEX", "AMS":
		return "AMS", true
	case "HK", "HKEX", "HONGKONG", "SEHK", "HKS":
		return "HKS", true
	case "JP", "JAPAN", "TSE", "JPX", "TKSE":
		return "TSE", true
	case "SH", "SHA", "SHAA", "SHS", "SSE", "SHANGHAI":
		return "SHS", true
	case "SZ", "SZA", "SZAA", "SZS", "SZSE", "SHENZHEN":
		return "SZS", true
	case "HNX", "HASE", "HANOI":
		return "HNX", true
	case "HSX", "VNSE", "HOCHIMINH":
		return "HSX", true
	default:
		return "", false
	}
}

func toKISProductTypeCode(market string) (string, bool, error) {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "", "KRX", "KOSPI", "KOSDAQ", "KNX", "KONEX", "NXT", "SOR":
		return "300", false, nil
	case "US", "US-NASDAQ", "NASDAQ", "NAS", "NASD":
		return "512", true, nil
	case "US-NYSE", "NYSE", "NYS":
		return "513", true, nil
	case "US-AMEX", "AMEX", "AMS":
		return "529", true, nil
	case "JP", "JAPAN", "TSE", "JPX":
		return "515", true, nil
	case "HK", "HKEX", "HONGKONG", "SEHK", "HKS":
		return "501", true, nil
	case "HKCNY":
		return "543", true, nil
	case "HKUSD":
		return "558", true, nil
	case "SH", "SHA", "SHAA", "SHS", "SSE", "SHANGHAI":
		return "551", true, nil
	case "SZ", "SZA", "SZAA", "SZS", "SZSE", "SHENZHEN":
		return "552", true, nil
	case "HNX", "HASE", "HANOI":
		return "507", true, nil
	case "HSX", "VNSE", "HOCHIMINH":
		return "508", true, nil
	default:
		return "", false, broker.ErrInvalidMarket
	}
}

func parseYN(v string) bool {
	return strings.EqualFold(strings.TrimSpace(v), "Y")
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func normalizeISIN(v string) string {
	s := strings.ToUpper(strings.TrimSpace(v))
	if len(s) != 12 {
		return ""
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return ""
		}
	}
	return s
}

func normalizeDomesticMarket(requested, marketID string) string {
	switch strings.ToUpper(strings.TrimSpace(marketID)) {
	case "STK":
		return "KOSPI"
	case "KSQ":
		return "KOSDAQ"
	case "KNX":
		return "KONEX"
	default:
		if requested == "" {
			return "KRX"
		}
		return requested
	}
}

func normalizeDomesticExchange(code string) string {
	switch strings.TrimSpace(code) {
	case "01", "02", "03":
		return "KRX"
	case "04":
		return "K-OTC"
	case "81":
		return "KRX-AH"
	default:
		return code
	}
}

func domesticAssetTypeFromSecurityGroup(group string) broker.AssetType {
	switch strings.ToUpper(strings.TrimSpace(group)) {
	case "EF", "EN", "FE", "MF", "RT", "SC", "TC":
		return broker.AssetFund
	default:
		return broker.AssetStock
	}
}

func domesticProductTypeFromSecurityGroup(group string) string {
	switch strings.ToUpper(strings.TrimSpace(group)) {
	case "EF":
		return "etf"
	case "EN":
		return "etn"
	case "EW":
		return "elw"
	case "OP":
		return "option"
	case "FU":
		return "futures"
	case "ST", "KN", "FS", "DR":
		return "stock"
	default:
		return "unknown"
	}
}

func overseasProductTypeFromCode(code string) string {
	switch strings.TrimSpace(code) {
	case "01":
		return "stock"
	case "02":
		return "warrant"
	case "03":
		return "etf"
	case "04":
		return "preferred"
	default:
		return "unknown"
	}
}

func toBrokerAssetType(market, productType string) broker.AssetType {
	if isDomesticMarketAlias(market) {
		switch productType {
		case "etf", "etn":
			return broker.AssetFund
		default:
			return broker.AssetStock
		}
	}

	switch productType {
	case "etf":
		return broker.AssetFund
	case "index":
		return broker.AssetOverseas
	default:
		return broker.AssetOverseas
	}
}

func isDomesticMarketAlias(market string) bool {
	switch strings.ToUpper(strings.TrimSpace(market)) {
	case "KRX", "KOSPI", "KOSDAQ", "KONEX", "KNX", "NXT":
		return true
	default:
		return false
	}
}

func applyOHLCVOptions(src []broker.OHLCV, opts broker.OHLCVOpts) ([]broker.OHLCV, error) {
	out := src
	if !opts.From.IsZero() || !opts.To.IsZero() {
		filtered := make([]broker.OHLCV, 0, len(src))
		for _, it := range src {
			if !opts.From.IsZero() && it.Timestamp.Before(opts.From) {
				continue
			}
			if !opts.To.IsZero() && it.Timestamp.After(opts.To) {
				continue
			}
			filtered = append(filtered, it)
		}
		out = filtered
	}

	interval := strings.ToLower(strings.TrimSpace(opts.Interval))
	switch interval {
	case "", "1d", "d", "day", "daily":
	case "1w", "w", "week", "weekly":
		out = aggregateOHLCVByPeriod(out, func(t time.Time) string {
			y, w := t.ISOWeek()
			return fmt.Sprintf("%04d-W%02d", y, w)
		})
	case "1mo", "mo", "month", "monthly":
		out = aggregateOHLCVByPeriod(out, func(t time.Time) string {
			y, m, _ := t.Date()
			return fmt.Sprintf("%04d-%02d", y, int(m))
		})
	default:
		return nil, fmt.Errorf("unsupported interval: %s", opts.Interval)
	}

	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[:opts.Limit]
	}
	return out, nil
}

func aggregateOHLCVByPeriod(src []broker.OHLCV, keyFn func(time.Time) string) []broker.OHLCV {
	if len(src) == 0 {
		return src
	}

	items := append([]broker.OHLCV(nil), src...)
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.Before(items[j].Timestamp) })

	result := make([]broker.OHLCV, 0, len(items))
	var curr broker.OHLCV
	var currKey string
	for i, it := range items {
		key := keyFn(it.Timestamp)
		if i == 0 || key != currKey {
			if i > 0 {
				result = append(result, curr)
			}
			currKey = key
			curr = it
			continue
		}
		if it.High > curr.High {
			curr.High = it.High
		}
		if it.Low < curr.Low || curr.Low == 0 {
			curr.Low = it.Low
		}
		curr.Close = it.Close
		curr.Volume += it.Volume
		curr.Timestamp = it.Timestamp
	}
	result = append(result, curr)

	sort.Slice(result, func(i, j int) bool { return result[i].Timestamp.After(result[j].Timestamp) })
	return result
}
