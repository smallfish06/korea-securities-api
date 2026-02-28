package kis

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// InquireOverseasPrice retrieves overseas stock price
// TR_ID: HHDFS00000300 (미국 주식 현재가)
// 거래소코드: NASDAQ(NAS), NYSE(NYS), AMEX(AMS)
func (c *Client) InquireOverseasPrice(ctx context.Context, exchangeCode, symbol string) (*OverseasPriceResponse, error) {
	// 실전/모의 TR_ID가 같을 수 있음 (문서 확인 필요)
	trID := "HHDFS00000300"

	// 거래소코드 매핑
	var excd string
	switch exchangeCode {
	case "NASDAQ":
		excd = "NAS"
	case "NYSE":
		excd = "NYS"
	case "AMEX":
		excd = "AMS"
	default:
		excd = exchangeCode // 그대로 사용
	}

	path := fmt.Sprintf("/uapi/overseas-price/v1/quotations/price?AUTH=&EXCD=%s&SYMB=%s",
		excd, symbol)

	var resp OverseasPriceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire overseas price: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireOverseasDailyChartPrice retrieves overseas day/week/month/year chart.
// TR_ID: FHKST03030100
func (c *Client) InquireOverseasDailyChartPrice(ctx context.Context, marketDiv, symbol, fromDate, toDate, periodDiv string) (*RawResponse, error) {
	if marketDiv == "" {
		marketDiv = "N"
	}
	if periodDiv == "" {
		periodDiv = "D"
	}
	q := url.Values{}
	q.Set("FID_COND_MRKT_DIV_CODE", marketDiv)
	q.Set("FID_INPUT_ISCD", symbol)
	q.Set("FID_INPUT_DATE_1", fromDate)
	q.Set("FID_INPUT_DATE_2", toDate)
	q.Set("FID_PERIOD_DIV_CODE", periodDiv)

	return c.getRaw(ctx,
		encodeQuery("/uapi/overseas-price/v1/quotations/inquire-daily-chartprice", q),
		"FHKST03030100",
	)
}

// InquireOverseasDailyPrice retrieves overseas period quotes.
// TR_ID: HHDFS76240000
func (c *Client) InquireOverseasDailyPrice(ctx context.Context, auth, exchangeCode, symbol, gubn, bymd, modp string) (*RawResponse, error) {
	if gubn == "" {
		gubn = "0"
	}
	if modp == "" {
		modp = "0"
	}
	q := url.Values{}
	q.Set("AUTH", auth)
	q.Set("EXCD", exchangeCode)
	q.Set("SYMB", symbol)
	q.Set("GUBN", gubn)
	q.Set("BYMD", bymd)
	q.Set("MODP", modp)

	return c.getRaw(ctx,
		encodeQuery("/uapi/overseas-price/v1/quotations/dailyprice", q),
		"HHDFS76240000",
	)
}

// InquireOverseasPriceDetail retrieves overseas orderbook/detail snapshot.
// TR_ID: HHDFS76200200
func (c *Client) InquireOverseasPriceDetail(ctx context.Context, auth, exchangeCode, symbol string) (*RawResponse, error) {
	q := url.Values{}
	q.Set("AUTH", auth)
	q.Set("EXCD", exchangeCode)
	q.Set("SYMB", symbol)

	return c.getRaw(ctx,
		encodeQuery("/uapi/overseas-price/v1/quotations/price-detail", q),
		"HHDFS76200200",
	)
}

// InquireOverseasTick retrieves overseas execution/tick details.
// TR_ID: HHDFS76200300
func (c *Client) InquireOverseasTick(ctx context.Context, exchangeCode, tday, symbol, auth, keyb string) (*RawResponse, error) {
	if tday == "" {
		tday = "0"
	}
	q := url.Values{}
	q.Set("EXCD", exchangeCode)
	q.Set("TDAY", tday)
	q.Set("SYMB", symbol)
	q.Set("AUTH", auth)
	q.Set("KEYB", keyb)

	return c.getRaw(ctx,
		encodeQuery("/uapi/overseas-price/v1/quotations/inquire-ccnl", q),
		"HHDFS76200300",
	)
}

// InquireOverseasUpdownRate retrieves overseas up/down rank.
// TR_ID: HHDFS76290000
func (c *Client) InquireOverseasUpdownRate(ctx context.Context, exchangeCode, nday, gubn, volRange, auth, keyb string) (*RawResponse, error) {
	if nday == "" {
		nday = "0"
	}
	if gubn == "" {
		gubn = "1"
	}
	if volRange == "" {
		volRange = "0"
	}
	q := url.Values{}
	q.Set("EXCD", exchangeCode)
	q.Set("NDAY", nday)
	q.Set("GUBN", gubn)
	q.Set("VOL_RANG", volRange)
	q.Set("AUTH", auth)
	q.Set("KEYB", keyb)

	return c.getRaw(ctx,
		encodeQuery("/uapi/overseas-stock/v1/ranking/updown-rate", q),
		"HHDFS76290000",
	)
}

// InquireOverseasTimeItemChartPrice retrieves overseas intraday time chart/ticks.
// TR_ID: HHDFS76950200
func (c *Client) InquireOverseasTimeItemChartPrice(ctx context.Context, auth, exchangeCode, symbol, nmin, pinc, next, nrec, fill, keyb string) (*RawResponse, error) {
	if nmin == "" {
		nmin = "1"
	}
	if pinc == "" {
		pinc = "1"
	}
	if next == "" {
		next = "0"
	}
	if nrec == "" {
		nrec = "120"
	}
	q := url.Values{}
	q.Set("AUTH", auth)
	q.Set("EXCD", exchangeCode)
	q.Set("SYMB", symbol)
	q.Set("NMIN", nmin)
	q.Set("PINC", pinc)
	q.Set("NEXT", next)
	q.Set("NREC", nrec)
	q.Set("FILL", fill)
	q.Set("KEYB", keyb)

	return c.getRaw(ctx,
		encodeQuery("/uapi/overseas-price/v1/quotations/inquire-time-itemchartprice", q),
		"HHDFS76950200",
	)
}

// IsSuccess checks if response is successful
func (r *OverseasPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

// OrderOverseas places an overseas stock order.
func (c *Client) OrderOverseas(
	ctx context.Context,
	accountNo, accountProductCode, exchangeCode, symbol string,
	quantity int,
	price float64,
	side string,
	ordDvsn string,
) (*OrderResponse, error) {
	ovrsExcg := normalizeOverseasExchangeCode(exchangeCode)
	trID, err := overseasOrderTRID(ovrsExcg, side, c.baseURL == BaseURLSandbox)
	if err != nil {
		return nil, err
	}
	if ordDvsn == "" {
		ordDvsn = "00"
	}

	sllType := ""
	if side == "sell" {
		sllType = "00"
	}

	req := OverseasOrderRequest{
		CANO:            accountNo,
		ACNT_PRDT_CD:    accountProductCode,
		OVRS_EXCG_CD:    ovrsExcg,
		PDNO:            symbol,
		ORD_QTY:         fmt.Sprintf("%d", quantity),
		OVRS_ORD_UNPR:   fmt.Sprintf("%.4f", price),
		SLL_TYPE:        sllType,
		ORD_SVR_DVSN_CD: "0",
		ORD_DVSN:        ordDvsn,
	}

	var resp OrderResponse
	if err := c.doRequest(ctx, "POST", "/uapi/overseas-stock/v1/trading/order", trID, req, &resp); err != nil {
		return nil, fmt.Errorf("order overseas: %w", err)
	}
	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCode)
	}

	return &resp, nil
}

// OrderOverseasRvseCncl places an overseas stock revise/cancel order.
func (c *Client) OrderOverseasRvseCncl(
	ctx context.Context,
	accountNo, accountProductCode, exchangeCode, symbol, originalOrderNo, rvseCnclDvsnCD string,
	quantity int,
	price float64,
) (*OrderResponse, error) {
	ovrsExcg := normalizeOverseasExchangeCode(exchangeCode)
	trID, err := overseasRvseCnclTRID(ovrsExcg, c.baseURL == BaseURLSandbox)
	if err != nil {
		return nil, err
	}

	req := OverseasOrderRvseCnclRequest{
		CANO:              accountNo,
		ACNT_PRDT_CD:      accountProductCode,
		OVRS_EXCG_CD:      ovrsExcg,
		PDNO:              symbol,
		ORGN_ODNO:         originalOrderNo,
		RVSE_CNCL_DVSN_CD: rvseCnclDvsnCD,
		ORD_QTY:           fmt.Sprintf("%d", quantity),
		OVRS_ORD_UNPR:     fmt.Sprintf("%.4f", price),
		ORD_SVR_DVSN_CD:   "0",
	}

	var resp OrderResponse
	if err := c.doRequest(ctx, "POST", "/uapi/overseas-stock/v1/trading/order-rvsecncl", trID, req, &resp); err != nil {
		return nil, fmt.Errorf("order overseas revise/cancel: %w", err)
	}
	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCode)
	}

	return &resp, nil
}

func normalizeOverseasExchangeCode(exchangeCode string) string {
	code := strings.ToUpper(strings.TrimSpace(exchangeCode))
	switch code {
	case "US", "NAS", "NASD", "NASDAQ", "US-NASDAQ":
		return "NASD"
	case "NYS", "NYSE", "US-NYSE":
		return "NYSE"
	case "AMS", "AMEX", "US-AMEX":
		return "AMEX"
	default:
		return code
	}
}

func overseasOrderTRID(exchangeCode, side string, sandbox bool) (string, error) {
	code := normalizeOverseasExchangeCode(exchangeCode)
	var trID string
	switch strings.ToLower(side) {
	case "buy":
		switch code {
		case "NASD", "NYSE", "AMEX":
			trID = "TTTT1002U"
		case "SEHK":
			trID = "TTTS1002U"
		case "SHAA":
			trID = "TTTS0202U"
		case "SZAA":
			trID = "TTTS0305U"
		case "TKSE":
			trID = "TTTS0308U"
		case "HASE", "VNSE":
			trID = "TTTS0311U"
		default:
			return "", fmt.Errorf("unsupported overseas exchange code: %s", exchangeCode)
		}
	case "sell":
		switch code {
		case "NASD", "NYSE", "AMEX":
			trID = "TTTT1006U"
		case "SEHK":
			trID = "TTTS1001U"
		case "SHAA":
			trID = "TTTS1005U"
		case "SZAA":
			trID = "TTTS0304U"
		case "TKSE":
			trID = "TTTS0307U"
		case "HASE", "VNSE":
			trID = "TTTS0310U"
		default:
			return "", fmt.Errorf("unsupported overseas exchange code: %s", exchangeCode)
		}
	default:
		return "", fmt.Errorf("unsupported order side: %s", side)
	}

	if sandbox && len(trID) > 0 {
		trID = "V" + trID[1:]
	}
	return trID, nil
}

func overseasRvseCnclTRID(exchangeCode string, sandbox bool) (string, error) {
	code := normalizeOverseasExchangeCode(exchangeCode)
	var trID string
	switch code {
	case "NASD", "NYSE", "AMEX":
		trID = "TTTT1004U"
	case "SEHK":
		trID = "TTTS1003U"
	case "SHAA":
		trID = "TTTS0302U"
	case "SZAA":
		trID = "TTTS0306U"
	case "TKSE":
		trID = "TTTS0309U"
	case "HASE", "VNSE":
		trID = "TTTS0312U"
	default:
		return "", fmt.Errorf("unsupported overseas exchange code: %s", exchangeCode)
	}

	if sandbox && len(trID) > 0 {
		trID = "V" + trID[1:]
	}
	return trID, nil
}
