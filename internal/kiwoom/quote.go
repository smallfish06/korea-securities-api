package kiwoom

import (
	"context"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// GetDomesticQuote fetches ka10001 and returns typed quote fields.
func (c *Client) GetDomesticQuote(ctx context.Context, symbol string) (*DomesticQuote, error) {
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
