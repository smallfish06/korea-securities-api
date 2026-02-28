package kis

import (
	"context"
	"fmt"
	"net/url"
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

// InquireAskingPriceExpCcn retrieves domestic orderbook/expected matching data.
// TR_ID: FHKST01010200
func (c *Client) InquireAskingPriceExpCcn(ctx context.Context, marketDiv, symbol string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "J"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", symbol)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-asking-price-exp-ccn", q),
		"FHKST01010200",
	)
}

// InquireCcnl retrieves domestic current execution/tick snapshot.
// TR_ID: FHKST01010300
func (c *Client) InquireCcnl(ctx context.Context, marketDiv, symbol string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "J"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", symbol)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-ccnl", q),
		"FHKST01010300",
	)
}

// InquireTimeItemConclusion retrieves intraday time-bucketed trade details.
// TR_ID: FHPST01060000
func (c *Client) InquireTimeItemConclusion(ctx context.Context, marketDiv, symbol, inputHour1 string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "J"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", symbol)
	q.Set("FID_INPUT_HOUR_1", inputHour1)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-time-itemconclusion", q),
		"FHPST01060000",
	)
}

// InquireMember retrieves broker/member level quote details.
// TR_ID: FHKST01010600
func (c *Client) InquireMember(ctx context.Context, marketDiv, symbol string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "J"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", symbol)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-member", q),
		"FHKST01010600",
	)
}

// InquireComponentStockPrice retrieves ETF/ETN component stocks.
// TR_ID: FHKST121600C0
func (c *Client) InquireComponentStockPrice(ctx context.Context, marketDiv, symbol, screenDiv string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "J"
	}
	if screenDiv == "" {
		screenDiv = "11216"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", symbol)
	q.Set("FID_COND_SCR_DIV_CODE", screenDiv)

	return c.getRaw(ctx,
		encodeQuery("/uapi/etfetn/v1/quotations/inquire-component-stock-price", q),
		"FHKST121600C0",
	)
}

// VolumeRankParams are request params for domestic volume rank API.
type VolumeRankParams struct {
	MarketDiv       string
	ScreenDiv       string
	InputISCD       string
	DivClsCode      string
	BlngClsCode     string
	TrgtClsCode     string
	TrgtExlsClsCode string
	InputPrice1     string
	InputPrice2     string
	VolCnt          string
	InputDate1      string
}

// InquireVolumeRank retrieves domestic volume rank.
// TR_ID: FHPST01710000
func (c *Client) InquireVolumeRank(ctx context.Context, p VolumeRankParams) (*RawResponse, error) {
	if p.MarketDiv == "" {
		p.MarketDiv = "J"
	}
	if p.ScreenDiv == "" {
		p.ScreenDiv = "20171"
	}
	if p.InputISCD == "" {
		p.InputISCD = "0000"
	}
	if p.DivClsCode == "" {
		p.DivClsCode = "0"
	}
	if p.BlngClsCode == "" {
		p.BlngClsCode = "0"
	}
	if p.TrgtClsCode == "" {
		p.TrgtClsCode = "111111111"
	}
	if p.TrgtExlsClsCode == "" {
		p.TrgtExlsClsCode = "0000000000"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", p.MarketDiv)
	q.Set("FID_COND_SCR_DIV_CODE", p.ScreenDiv)
	q.Set("FID_INPUT_ISCD", p.InputISCD)
	q.Set("FID_DIV_CLS_CODE", p.DivClsCode)
	q.Set("FID_BLNG_CLS_CODE", p.BlngClsCode)
	q.Set("FID_TRGT_CLS_CODE", p.TrgtClsCode)
	q.Set("FID_TRGT_EXLS_CLS_CODE", p.TrgtExlsClsCode)
	q.Set("FID_INPUT_PRICE_1", p.InputPrice1)
	q.Set("FID_INPUT_PRICE_2", p.InputPrice2)
	q.Set("FID_VOL_CNT", p.VolCnt)
	q.Set("FID_INPUT_DATE_1", p.InputDate1)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/volume-rank", q),
		"FHPST01710000",
	)
}

// MarketCapRankParams are request params for market-cap rank API.
type MarketCapRankParams struct {
	InputPrice2     string
	MarketDiv       string
	ScreenDiv       string
	DivClsCode      string
	InputISCD       string
	TrgtClsCode     string
	TrgtExlsClsCode string
	InputPrice1     string
	VolCnt          string
}

// InquireMarketCapRank retrieves market-cap rank.
// TR_ID: FHPST01740000
func (c *Client) InquireMarketCapRank(ctx context.Context, p MarketCapRankParams) (*RawResponse, error) {
	if p.MarketDiv == "" {
		p.MarketDiv = "J"
	}
	if p.ScreenDiv == "" {
		p.ScreenDiv = "20174"
	}
	if p.DivClsCode == "" {
		p.DivClsCode = "0"
	}
	if p.InputISCD == "" {
		p.InputISCD = "0000"
	}
	if p.TrgtClsCode == "" {
		p.TrgtClsCode = "0"
	}
	if p.TrgtExlsClsCode == "" {
		p.TrgtExlsClsCode = "0"
	}

	q := url.Values{}
	q.Set("fid_input_price_2", p.InputPrice2)
	q.Set("fid_cond_mrkt_div_code", p.MarketDiv)
	q.Set("fid_cond_scr_div_code", p.ScreenDiv)
	q.Set("fid_div_cls_code", p.DivClsCode)
	q.Set("fid_input_iscd", p.InputISCD)
	q.Set("fid_trgt_cls_code", p.TrgtClsCode)
	q.Set("fid_trgt_exls_cls_code", p.TrgtExlsClsCode)
	q.Set("fid_input_price_1", p.InputPrice1)
	q.Set("fid_vol_cnt", p.VolCnt)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/ranking/market-cap", q),
		"FHPST01740000",
	)
}

// InquireIndexPrice retrieves domestic index current quote.
// TR_ID: FHPUP02100000
func (c *Client) InquireIndexPrice(ctx context.Context, marketDiv, indexCode string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "U"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", indexCode)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-index-price", q),
		"FHPUP02100000",
	)
}

// InquireIndexDailyPrice retrieves domestic index daily series.
// TR_ID: FHPUP02120000
func (c *Client) InquireIndexDailyPrice(ctx context.Context, periodDiv, marketDiv, indexCode, inputDate1 string) (*RawResponse, error) {
	if periodDiv == "" {
		periodDiv = "D"
	}
	if marketDiv == "" {
		marketDiv = "U"
	}
	q := url.Values{}
	q.Set("FID_PERIOD_DIV_CODE", periodDiv)
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", indexCode)
	q.Set("FID_INPUT_DATE_1", inputDate1)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-index-daily-price", q),
		"FHPUP02120000",
	)
}

// InquireDailyIndexChartPrice retrieves sector/index period chart.
// TR_ID: FHKUP03500100
func (c *Client) InquireDailyIndexChartPrice(ctx context.Context, marketDiv, indexCode, fromDate, toDate, periodDiv string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "U"
	}
	if periodDiv == "" {
		periodDiv = "D"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", indexCode)
	q.Set("FID_INPUT_DATE_1", fromDate)
	q.Set("FID_INPUT_DATE_2", toDate)
	q.Set("FID_PERIOD_DIV_CODE", periodDiv)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/quotations/inquire-daily-indexchartprice", q),
		"FHKUP03500100",
	)
}

// InquireDividend retrieves KSD dividend schedule.
// TR_ID: HHKDB669102C0
func (c *Client) InquireDividend(ctx context.Context, cts, gb1, fromDate, toDate, shortCode, highGb string) (*RawResponse, error) {
	q := url.Values{}
	q.Set("CTS", cts)
	q.Set("GB1", gb1)
	q.Set("F_DT", fromDate)
	q.Set("T_DT", toDate)
	q.Set("SHT_CD", shortCode)
	q.Set("HIGH_GB", highGb)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/ksdinfo/dividend", q),
		"HHKDB669102C0",
	)
}

// InquireFinancialRatio retrieves domestic financial ratios.
// TR_ID: FHKST66430300
func (c *Client) InquireFinancialRatio(ctx context.Context, divClsCode, marketDiv, symbol string) (*RawResponse, error) {
	if divClsCode == "" {
		divClsCode = "0"
	}
	if marketDiv == "" {
		marketDiv = "J"
	}
	q := url.Values{}
	q.Set("FID_DIV_CLS_CODE", divClsCode)
	q.Set("fid_cond_mrkt_div_code", marketDiv)
	q.Set("fid_input_iscd", symbol)

	return c.getRaw(ctx,
		encodeQuery("/uapi/domestic-stock/v1/finance/financial-ratio", q),
		"FHKST66430300",
	)
}

// IsSuccess checks if response is successful
func (r *StockPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *StockDailyPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

// Rows returns normalized daily rows regardless of output/output1 variant.
func (r *StockDailyPriceResponse) Rows() []StockDailyPriceOutput {
	if len(r.Output) > 0 {
		return r.Output
	}
	return r.Output1
}

func (r *StockBalanceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *StockRvseCnclResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
