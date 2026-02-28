package kis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// QuoteResponse represents KIS quote response
type QuoteResponse struct {
	RetCode string `json:"rt_cd"`
	MsgCode string `json:"msg_cd"`
	Msg1    string `json:"msg1"`
	Output  struct {
		StockCode      string `json:"stck_shrn_iscd"` // 종목코드
		PrdyVrss       string `json:"prdy_vrss"`      // 전일대비
		PrdyVrssSign   string `json:"prdy_vrss_sign"` // 전일대비부호
		PrdyCtrt       string `json:"prdy_ctrt"`      // 전일대비율
		StckPrpr       string `json:"stck_prpr"`      // 현재가
		StckOprc       string `json:"stck_oprc"`      // 시가
		StckHgpr       string `json:"stck_hgpr"`      // 고가
		StckLwpr       string `json:"stck_lwpr"`      // 저가
		StckMxpr       string `json:"stck_mxpr"`      // 상한가
		StckLlam       string `json:"stck_llam"`      // 하한가
		AccTrVol       string `json:"acml_vol"`       // 누적거래량
		AccTrPbmn      string `json:"acml_tr_pbmn"`   // 누적거래대금
		StckSdpr       string `json:"stck_sdpr"`      // 기준가
		PrdyVol        string `json:"prdy_vol"`       // 전일거래량
		StckFcam       string `json:"stck_fcam"`      // 액면가
		AscnLmtPriceRt string `json:"ascn_lmt_price"` // 상승제한가
		DscnLmtPriceRt string `json:"dscn_lmt_price"` // 하락제한가
		HtsKorIsnm     string `json:"hts_kor_isnm"`   // 종목명
	} `json:"output"`
}

// GetQuote retrieves a quote for the given market and symbol
func (c *Client) GetQuote(ctx context.Context, market, symbol string) (*broker.Quote, error) {
	// KIS API 경로: /uapi/domestic-stock/v1/quotations/inquire-price
	// Query params: FID_COND_MRKT_DIV_CODE=J&FID_INPUT_ISCD={symbol}
	path := fmt.Sprintf("/uapi/domestic-stock/v1/quotations/inquire-price?FID_COND_MRKT_DIV_CODE=J&FID_INPUT_ISCD=%s", symbol)

	var resp QuoteResponse
	if err := c.doRequest(ctx, "GET", path, "FHKST01010100", nil, &resp); err != nil {
		return nil, fmt.Errorf("get quote: %w", err)
	}

	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS error: %s - %s", resp.MsgCode, resp.Msg1)
	}

	price, _ := strconv.ParseFloat(resp.Output.StckPrpr, 64)
	open, _ := strconv.ParseFloat(resp.Output.StckOprc, 64)
	high, _ := strconv.ParseFloat(resp.Output.StckHgpr, 64)
	low, _ := strconv.ParseFloat(resp.Output.StckLwpr, 64)
	change, _ := strconv.ParseFloat(resp.Output.PrdyVrss, 64)
	changeRate, _ := strconv.ParseFloat(resp.Output.PrdyCtrt, 64)
	volume, _ := strconv.ParseInt(resp.Output.AccTrVol, 10, 64)
	turnover, _ := strconv.ParseFloat(resp.Output.AccTrPbmn, 64)
	upperLimit, _ := strconv.ParseFloat(resp.Output.StckMxpr, 64)
	lowerLimit, _ := strconv.ParseFloat(resp.Output.StckLlam, 64)
	prevClose := price - change

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
		ChangeRate: changeRate,
		Volume:     volume,
		Turnover:   turnover,
		UpperLimit: upperLimit,
		LowerLimit: lowerLimit,
		Timestamp:  time.Now(),
	}, nil
}

// OHLCVResponse represents KIS OHLCV response
type OHLCVResponse struct {
	RetCode string `json:"rt_cd"`
	MsgCode string `json:"msg_cd"`
	Msg1    string `json:"msg1"`
	Output  []struct {
		StckBsopDate string `json:"stck_bsop_date"` // 영업일자
		StckOprc     string `json:"stck_oprc"`      // 시가
		StckHgpr     string `json:"stck_hgpr"`      // 고가
		StckLwpr     string `json:"stck_lwpr"`      // 저가
		StckClpr     string `json:"stck_clpr"`      // 종가
		AccTrVol     string `json:"acml_vol"`       // 누적거래량
	} `json:"output"`
}

// GetOHLCV retrieves OHLCV data for the given market and symbol
func (c *Client) GetOHLCV(ctx context.Context, market, symbol string, opts broker.OHLCVOpts) ([]broker.OHLCV, error) {
	// KIS API: /uapi/domestic-stock/v1/quotations/inquire-daily-itemchartprice
	// Query: FID_COND_MRKT_DIV_CODE=J&FID_INPUT_ISCD={symbol}&FID_PERIOD_DIV_CODE=D&FID_ORG_ADJ_PRC=0
	path := fmt.Sprintf("/uapi/domestic-stock/v1/quotations/inquire-daily-itemchartprice?FID_COND_MRKT_DIV_CODE=J&FID_INPUT_ISCD=%s&FID_PERIOD_DIV_CODE=D&FID_ORG_ADJ_PRC=0", symbol)

	var resp OHLCVResponse
	if err := c.doRequest(ctx, "GET", path, "FHKST01010400", nil, &resp); err != nil {
		return nil, fmt.Errorf("get ohlcv: %w", err)
	}

	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS error: %s - %s", resp.MsgCode, resp.Msg1)
	}

	result := make([]broker.OHLCV, 0, len(resp.Output))
	for _, item := range resp.Output {
		timestamp, _ := time.Parse("20060102", item.StckBsopDate)
		open, _ := strconv.ParseFloat(item.StckOprc, 64)
		high, _ := strconv.ParseFloat(item.StckHgpr, 64)
		low, _ := strconv.ParseFloat(item.StckLwpr, 64)
		close, _ := strconv.ParseFloat(item.StckClpr, 64)
		volume, _ := strconv.ParseInt(item.AccTrVol, 10, 64)

		result = append(result, broker.OHLCV{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return result, nil
}
