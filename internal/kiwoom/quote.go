package kiwoom

import (
	"context"
	"strings"
	"time"

	"github.com/smallfish06/krsec/pkg/broker"
)

// InquirePrice fetches ka10001 and returns typed quote fields.
func (c *Client) InquirePrice(ctx context.Context, symbol string) (*DomesticQuote, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}

	res, err := c.call(ctx, endpointDomesticQuote, map[string]interface{}{
		"stk_cd": symbol,
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	q := &DomesticQuote{
		Symbol:     normalizeSymbolCode(asString(res.Body["stk_cd"])),
		Name:       asString(res.Body["stk_nm"]),
		Price:      asFloat64(res.Body["cur_prc"]),
		Open:       asFloat64(res.Body["open_pric"]),
		High:       asFloat64(res.Body["high_pric"]),
		Low:        asFloat64(res.Body["low_pric"]),
		BasePrice:  asFloat64(res.Body["base_pric"]),
		UpperLimit: asFloat64(res.Body["upl_pric"]),
		LowerLimit: asFloat64(res.Body["lst_pric"]),
		Change:     asFloat64(res.Body["pred_pre"]),
		ChangeRate: asFloat64(res.Body["flu_rt"]),
		Volume:     asInt64(res.Body["trde_qty"]),
		ReturnMsg:  asString(res.Body["return_msg"]),
		ReturnCode: parseReturnCode(res.Body["return_code"]),
	}
	if q.Symbol == "" {
		q.Symbol = symbol
	}
	return q, nil
}

// InquireOrderBook fetches domestic orderbook/remaining sizes via ka10004.
func (c *Client) InquireOrderBook(ctx context.Context, symbol string) (map[string]interface{}, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	return c.callRaw(ctx, endpointDomesticOrderBook, map[string]interface{}{"stk_cd": symbol})
}

// InquireExecutionInfo fetches domestic execution info via ka10003.
func (c *Client) InquireExecutionInfo(ctx context.Context, symbol string) (map[string]interface{}, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	return c.callRaw(ctx, endpointDomesticExecutionInfo, map[string]interface{}{"stk_cd": symbol})
}

// InquireVolumeRank fetches domestic volume ranking via ka10030.
func (c *Client) InquireVolumeRank(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	payload := cloneBody(body)
	setDefaultPayload(payload, "stex_tp", "0")
	setDefaultPayload(payload, "mrkt_tp", "000")
	setDefaultPayload(payload, "sort_tp", "1")
	setDefaultPayload(payload, "mang_stk_incls", "0")
	setDefaultPayload(payload, "crd_tp", "0")
	setDefaultPayload(payload, "trde_qty_tp", "0")
	setDefaultPayload(payload, "pric_tp", "0")
	setDefaultPayload(payload, "trde_prica_tp", "0")
	setDefaultPayload(payload, "mrkt_open_tp", "0")
	return c.callRaw(ctx, endpointVolumeRank, payload)
}

// InquireChangeRateRank fetches domestic change-rate ranking via ka10027.
func (c *Client) InquireChangeRateRank(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	payload := cloneBody(body)
	setDefaultPayload(payload, "stex_tp", "0")
	setDefaultPayload(payload, "mrkt_tp", "000")
	setDefaultPayload(payload, "sort_tp", "1")
	setDefaultPayload(payload, "trde_qty_cnd", "0")
	setDefaultPayload(payload, "stk_cnd", "0")
	setDefaultPayload(payload, "crd_cnd", "0")
	setDefaultPayload(payload, "updown_incls", "1")
	setDefaultPayload(payload, "pric_cnd", "0")
	setDefaultPayload(payload, "trde_prica_cnd", "0")
	return c.callRaw(ctx, endpointChangeRateRank, payload)
}

// InquireInvestorByStock fetches investor trend by stock via ka10059.
func (c *Client) InquireInvestorByStock(ctx context.Context, symbol string, body map[string]interface{}) (map[string]interface{}, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	payload := cloneBody(body)
	setDefaultPayload(payload, "dt", time.Now().Format("20060102"))
	setDefaultPayload(payload, "trde_tp", "0")
	setDefaultPayload(payload, "amt_qty_tp", "1")
	setDefaultPayload(payload, "unit_tp", "1")
	payload["stk_cd"] = symbol
	return c.callRaw(ctx, endpointInvestorByStock, payload)
}

// InquireSectorCurrent fetches sector current quote via ka20001.
func (c *Client) InquireSectorCurrent(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	payload := cloneBody(body)
	setDefaultPayload(payload, "mrkt_tp", "000")
	upcode := strings.TrimSpace(asString(payload["upcode"]))
	if strings.TrimSpace(asString(payload["inds_cd"])) == "" && upcode != "" {
		payload["inds_cd"] = upcode
	}
	return c.callRaw(ctx, endpointSectorCurrent, payload)
}

// InquireSectorByPrice fetches sector-by-price data via ka20002.
func (c *Client) InquireSectorByPrice(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	payload := cloneBody(body)
	setDefaultPayload(payload, "mrkt_tp", "000")
	setDefaultPayload(payload, "stex_tp", "0")
	upcode := strings.TrimSpace(asString(payload["upcode"]))
	if strings.TrimSpace(asString(payload["inds_cd"])) == "" && upcode != "" {
		payload["inds_cd"] = upcode
	}
	return c.callRaw(ctx, endpointSectorByPrice, payload)
}

// InquireELWDetail fetches ELW/ETF-related detail via ka30012.
func (c *Client) InquireELWDetail(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	return c.callRaw(ctx, endpointELWDetail, body)
}

func setDefaultPayload(payload map[string]interface{}, key, value string) {
	if strings.TrimSpace(asString(payload[key])) == "" {
		payload[key] = value
	}
}
