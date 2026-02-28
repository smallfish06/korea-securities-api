package kiwoom

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// PlaceStockOrder places kt10000 (buy) or kt10001 (sell).
func (c *Client) PlaceStockOrder(ctx context.Context, req PlaceStockOrderRequest) (*OrderAck, error) {
	req.Symbol = normalizeSymbolCode(req.Symbol)
	if req.Symbol == "" || req.Quantity <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}
	req.Exchange = strings.ToUpper(strings.TrimSpace(req.Exchange))
	if req.Exchange == "" {
		req.Exchange = "KRX"
	}
	req.TradeType = strings.TrimSpace(req.TradeType)
	if req.TradeType == "" {
		return nil, broker.ErrInvalidOrderRequest
	}

	var apiID string
	switch req.Side {
	case StockOrderSideBuy:
		apiID = "kt10000"
	case StockOrderSideSell:
		apiID = "kt10001"
	default:
		return nil, broker.ErrInvalidOrderRequest
	}

	res, err := c.call(ctx, apiID, map[string]interface{}{
		"dmst_stex_tp": req.Exchange,
		"stk_cd":       req.Symbol,
		"ord_qty":      fmt.Sprintf("%d", req.Quantity),
		"ord_uv":       strings.TrimSpace(req.OrderPrice),
		"trde_tp":      req.TradeType,
		"cond_uv":      strings.TrimSpace(req.ConditionPrice),
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	orderID := asString(res.Body["ord_no"])
	if orderID == "" {
		return nil, fmt.Errorf("missing order id in kiwoom response")
	}

	return &OrderAck{
		OrderNumber: orderID,
		ReturnMsg:   asString(res.Body["return_msg"]),
		ReturnCode:  parseReturnCode(res.Body["return_code"]),
	}, nil
}

// CancelStockOrder cancels order through kt10003.
func (c *Client) CancelStockOrder(ctx context.Context, req CancelStockOrderRequest) (*OrderAck, error) {
	req.Exchange = strings.ToUpper(strings.TrimSpace(req.Exchange))
	if req.Exchange == "" {
		req.Exchange = "KRX"
	}
	req.Symbol = normalizeSymbolCode(req.Symbol)
	req.OriginalID = strings.TrimSpace(req.OriginalID)
	if req.Symbol == "" || req.OriginalID == "" || req.CancelQty <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}

	res, err := c.call(ctx, "kt10003", map[string]interface{}{
		"dmst_stex_tp": req.Exchange,
		"orig_ord_no":  req.OriginalID,
		"stk_cd":       req.Symbol,
		"cncl_qty":     fmt.Sprintf("%d", req.CancelQty),
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	return &OrderAck{
		OrderNumber: asString(res.Body["ord_no"]),
		ReturnMsg:   asString(res.Body["return_msg"]),
		ReturnCode:  parseReturnCode(res.Body["return_code"]),
	}, nil
}

// ModifyStockOrder modifies order through kt10002.
func (c *Client) ModifyStockOrder(ctx context.Context, req ModifyStockOrderRequest) (*OrderAck, error) {
	req.Exchange = strings.ToUpper(strings.TrimSpace(req.Exchange))
	if req.Exchange == "" {
		req.Exchange = "KRX"
	}
	req.Symbol = normalizeSymbolCode(req.Symbol)
	req.OriginalID = strings.TrimSpace(req.OriginalID)
	if req.Symbol == "" || req.OriginalID == "" || req.ModifyQty <= 0 || strings.TrimSpace(req.ModifyPrice) == "" {
		return nil, broker.ErrInvalidOrderRequest
	}

	res, err := c.call(ctx, "kt10002", map[string]interface{}{
		"dmst_stex_tp": req.Exchange,
		"orig_ord_no":  req.OriginalID,
		"stk_cd":       req.Symbol,
		"mdfy_qty":     fmt.Sprintf("%d", req.ModifyQty),
		"mdfy_uv":      strings.TrimSpace(req.ModifyPrice),
		"mdfy_cond_uv": strings.TrimSpace(req.ConditionPrice),
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	return &OrderAck{
		OrderNumber: asString(res.Body["ord_no"]),
		ReturnMsg:   asString(res.Body["return_msg"]),
		ReturnCode:  parseReturnCode(res.Body["return_code"]),
	}, nil
}
