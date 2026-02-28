package kiwoom

import (
	"context"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// GetInstrumentInfo fetches ka10100.
func (c *Client) GetInstrumentInfo(ctx context.Context, symbol string) (*InstrumentInfo, error) {
	symbol = normalizeSymbolCode(symbol)
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}

	res, err := c.call(ctx, endpointInstrumentInfo, map[string]interface{}{
		"stk_cd": symbol,
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	info := &InstrumentInfo{
		Code:       normalizeSymbolCode(asString(res.Body["code"])),
		Name:       asString(res.Body["name"]),
		ListCount:  asInt64(res.Body["listCount"]),
		RegDay:     asString(res.Body["regDay"]),
		State:      asString(res.Body["state"]),
		MarketCode: asString(res.Body["marketCode"]),
		MarketName: asString(res.Body["marketName"]),
		SectorName: asString(res.Body["upName"]),
		ReturnMsg:  asString(res.Body["return_msg"]),
		ReturnCode: parseReturnCode(res.Body["return_code"]),
	}
	if info.Code == "" {
		return nil, broker.ErrInstrumentNotFound
	}
	return info, nil
}
