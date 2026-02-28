package kiwoom

import (
	"context"
	"strings"
	"time"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// GetDailyChart fetches daily candles from ka10081.
func (c *Client) GetDailyChart(ctx context.Context, symbol, baseDate string) ([]ChartCandle, error) {
	return c.getChart(ctx, endpointDailyChart, symbol, baseDate)
}

// GetWeeklyChart fetches weekly candles from ka10082.
func (c *Client) GetWeeklyChart(ctx context.Context, symbol, baseDate string) ([]ChartCandle, error) {
	return c.getChart(ctx, endpointWeeklyChart, symbol, baseDate)
}

// GetMonthlyChart fetches monthly candles from ka10083.
func (c *Client) GetMonthlyChart(ctx context.Context, symbol, baseDate string) ([]ChartCandle, error) {
	return c.getChart(ctx, endpointMonthlyChart, symbol, baseDate)
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
