package kis

import (
	"context"
	"fmt"
)

// InquirePrice retrieves current stock price
// TR_ID: FHKST01010100 (실전투자), VHKST01010300 (모의투자)
func (c *Client) InquirePrice(ctx context.Context, market, symbol string) (*StockPriceResponse, error) {
	trID := "FHKST01010100"
	if c.baseURL == BaseURLSandbox {
		trID = "VHKST01010300"
	}

	// FID 조건 매핑
	fidCondMrktDivCode := "J" // J: 주식, ETF, ETN
	if market == "KOSDAQ" {
		fidCondMrktDivCode = "Q"
	}

	path := fmt.Sprintf("/uapi/domestic-stock/v1/quotations/inquire-price?fid_cond_mrkt_div_code=%s&fid_input_iscd=%s",
		fidCondMrktDivCode, symbol)

	var resp StockPriceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire price: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireDailyPrice retrieves daily OHLCV data
// TR_ID: FHKST01010400 (실전투자), VHKST01010400 (모의투자)
func (c *Client) InquireDailyPrice(ctx context.Context, market, symbol string, startDate, endDate string, adjustPrice bool) (*StockDailyPriceResponse, error) {
	trID := "FHKST01010400"
	if c.baseURL == BaseURLSandbox {
		trID = "VHKST01010400"
	}

	fidCondMrktDivCode := "J"
	if market == "KOSDAQ" {
		fidCondMrktDivCode = "Q"
	}

	// 수정주가 여부 (0: 수정주가반영, 1: 수정주가미반영)
	fidOrgAdjPrc := "0"
	if !adjustPrice {
		fidOrgAdjPrc = "1"
	}

	path := fmt.Sprintf("/uapi/domestic-stock/v1/quotations/inquire-daily-price?fid_cond_mrkt_div_code=%s&fid_input_iscd=%s&fid_period_div_code=D&fid_org_adj_prc=%s",
		fidCondMrktDivCode, symbol, fidOrgAdjPrc)

	var resp StockDailyPriceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire daily price: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireBalance retrieves stock account balance
// TR_ID: TTTC8434R (실전투자), VTTC8434R (모의투자)
func (c *Client) InquireBalance(ctx context.Context, accountNo, accountProductCode string) (*StockBalanceResponse, error) {
	trID := "TTTC8434R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTC8434R"
	}

	// 계좌번호는 "12345678-01" 형식에서 앞부분이 CANO, 뒷부분이 ACNT_PRDT_CD
	cano := accountNo
	acntPrdtCd := accountProductCode
	if accountProductCode == "" {
		acntPrdtCd = "01" // 기본값
	}

	path := fmt.Sprintf("/uapi/domestic-stock/v1/trading/inquire-balance?CANO=%s&ACNT_PRDT_CD=%s&AFHR_FLPR_YN=N&INQR_DVSN=01&UNPR_DVSN=01&FUND_STTL_ICLD_YN=N&FNCG_AMT_AUTO_RDPT_YN=N&OFL_YN=&PRCS_DVSN=00&CTX_AREA_FK100=&CTX_AREA_NK100=",
		cano, acntPrdtCd)

	var resp StockBalanceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire balance: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquirePossibleRvseCncl retrieves cancellable/modifiable stock orders.
// TR_ID: TTTC8036R (legacy), VTTC8036R (legacy mock)
func (c *Client) InquirePossibleRvseCncl(ctx context.Context, accountNo, accountProductCode string) (*StockRvseCnclResponse, error) {
	trID := "TTTC8036R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTC8036R"
	}

	cano := accountNo
	acntPrdtCd := accountProductCode
	if accountProductCode == "" {
		acntPrdtCd = "01"
	}

	// INQR_DVSN_1: 0(주문), INQR_DVSN_2: 0(전체)
	path := fmt.Sprintf("/uapi/domestic-stock/v1/trading/inquire-psbl-rvsecncl?CANO=%s&ACNT_PRDT_CD=%s&CTX_AREA_FK100=&CTX_AREA_NK100=&INQR_DVSN_1=0&INQR_DVSN_2=0",
		cano, acntPrdtCd)

	var resp StockRvseCnclResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire possible revise/cancel: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// OrderCash places a cash order (buy/sell)
// TR_ID: TTTC0802U (실전 매수), VTTC0802U (모의 매수)
// TR_ID: TTTC0801U (실전 매도), VTTC0801U (모의 매도)
func (c *Client) OrderCash(ctx context.Context, accountNo, accountProductCode, symbol string, orderType string, quantity int, price int, side string, exchangeID string) (*OrderResponse, error) {
	// side: "buy" or "sell"
	trID := "TTTC0802U" // 매수
	if side == "sell" {
		trID = "TTTC0801U" // 매도
	}

	if c.baseURL == BaseURLSandbox {
		if side == "sell" {
			trID = "VTTC0801U"
		} else {
			trID = "VTTC0802U"
		}
	}

	// 주문구분: 00(지정가), 01(시장가)
	ordDvsn := "00"
	if orderType == "market" {
		ordDvsn = "01"
	}
	if exchangeID == "" {
		exchangeID = "KRX"
	}

	req := OrderRequest{
		CANO:         accountNo,
		ACNT_PRDT_CD: accountProductCode,
		PDNO:         symbol,
		ORD_DVSN:     ordDvsn,
		ORD_QTY:      fmt.Sprintf("%d", quantity),
		ORD_UNPR:     fmt.Sprintf("%d", price),
		EXCG_ID_DVSN: exchangeID,
		SLL_TYPE:     "01",
	}

	var resp OrderResponse
	if err := c.doRequest(ctx, "POST", "/uapi/domestic-stock/v1/trading/order-cash", trID, req, &resp); err != nil {
		return nil, fmt.Errorf("order cash: %w", err)
	}

	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCode)
	}

	return &resp, nil
}

// OrderRvseCncl places a domestic stock order revise/cancel request.
// RVSE_CNCL_DVSN_CD: 01=revise, 02=cancel
func (c *Client) OrderRvseCncl(
	ctx context.Context,
	accountNo, accountProductCode, orderOrgNo, originalOrderNo, orderDvsn, rvseCnclDvsnCD string,
	orderQty, orderPrice int,
	qtyAll bool,
	exchangeID string,
) (*OrderResponse, error) {
	trID := "TTTC0803U"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTC0803U"
	}

	if exchangeID == "" {
		exchangeID = "KRX"
	}

	req := OrderRvseCnclRequest{
		CANO:           accountNo,
		ACNT_PRDT_CD:   accountProductCode,
		KRXFwdOrdOrgNo: orderOrgNo,
		OrgnODNo:       originalOrderNo,
		OrdDvsn:        orderDvsn,
		RvseCnclDvsnCD: rvseCnclDvsnCD,
		OrdQty:         fmt.Sprintf("%d", orderQty),
		OrdUNPR:        fmt.Sprintf("%d", orderPrice),
		QtyAllOrdYN:    "N",
		ExcgIDDvsnCD:   exchangeID,
	}
	if qtyAll {
		req.QtyAllOrdYN = "Y"
	}

	var resp OrderResponse
	if err := c.doRequest(ctx, "POST", "/uapi/domestic-stock/v1/trading/order-rvsecncl", trID, req, &resp); err != nil {
		return nil, fmt.Errorf("order revise/cancel: %w", err)
	}
	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCode)
	}

	return &resp, nil
}

// IsSuccess checks if response is successful
func (r *StockPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *StockDailyPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *StockBalanceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *StockRvseCnclResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
