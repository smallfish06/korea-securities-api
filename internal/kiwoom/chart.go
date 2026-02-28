package kiwoom

import (
	"context"
	"strings"
	"time"

	"github.com/smallfish06/krsec/pkg/broker"
)

// InquireDailyPrice fetches daily candles from ka10081.
func (c *Client) InquireDailyPrice(ctx context.Context, symbol, baseDate string) ([]ChartCandle, error) {
	return c.getChart(ctx, endpointDailyChart, symbol, baseDate)
}

// InquireWeeklyPrice fetches weekly candles from ka10082.
func (c *Client) InquireWeeklyPrice(ctx context.Context, symbol, baseDate string) ([]ChartCandle, error) {
	return c.getChart(ctx, endpointWeeklyChart, symbol, baseDate)
}

// InquireMonthlyPrice fetches monthly candles from ka10083.
func (c *Client) InquireMonthlyPrice(ctx context.Context, symbol, baseDate string) ([]ChartCandle, error) {
	return c.getChart(ctx, endpointMonthlyChart, symbol, baseDate)
}

// InquireTickChart fetches domestic tick chart via ka10079.
func (c *Client) InquireTickChart(ctx context.Context, symbol, baseDate string) (map[string]interface{}, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	baseDate = strings.TrimSpace(baseDate)
	if baseDate == "" {
		baseDate = time.Now().Format("20060102")
	}
	body := map[string]interface{}{
		"stk_cd":       symbol,
		"base_dt":      baseDate,
		"upd_stkpc_tp": "1",
		"tic_scope":    "1",
	}
	return c.callRaw(ctx, endpointTickChart, body)
}

// InquireInvestorByStockChart fetches investor trend chart via ka10060.
func (c *Client) InquireInvestorByStockChart(ctx context.Context, symbol string, body map[string]interface{}) (map[string]interface{}, error) {
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
	return c.callRaw(ctx, endpointInvestorByStockChart, payload)
}

func (c *Client) getChart(ctx context.Context, endpoint endpointSpec, symbol, baseDate string) ([]ChartCandle, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	baseDate = strings.TrimSpace(baseDate)
	if baseDate == "" {
		baseDate = time.Now().Format("20060102")
	}

	res, err := c.call(ctx, endpoint, map[string]interface{}{
		"stk_cd":       symbol,
		"base_dt":      baseDate,
		"upd_stkpc_tp": "1",
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	rows := firstObjectArray(res.Body,
		"stk_dt_pole_chart_qry",
		"stk_stk_pole_chart_qry",
		"stk_mth_pole_chart_qry",
	)
	if len(rows) == 0 {
		return []ChartCandle{}, nil
	}

	candles := make([]ChartCandle, 0, len(rows))
	for _, row := range rows {
		dt, err := parseDateYYYYMMDD(asString(row["dt"]))
		if err != nil {
			continue
		}
		candles = append(candles, ChartCandle{
			Date:   dt,
			Open:   asFloat64(row["open_pric"]),
			High:   asFloat64(row["high_pric"]),
			Low:    asFloat64(row["low_pric"]),
			Close:  asFloat64(row["cur_prc"]),
			Volume: asInt64(row["trde_qty"]),
		})
	}
	return candles, nil
}
