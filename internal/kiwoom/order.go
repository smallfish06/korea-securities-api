package kiwoom

import (
	"context"
	"fmt"
	"strings"

	kiwoomspecs "github.com/smallfish06/krsec/internal/kiwoom/specs"
	"github.com/smallfish06/krsec/pkg/broker"
)

// PlaceStockOrder places kt10000 (buy) or kt10001 (sell).
func (c *Client) PlaceStockOrder(
	ctx context.Context,
	side StockOrderSide,
	req kiwoomspecs.KiwoomApiDostkOrdrKt10000Request,
) (*kiwoomspecs.KiwoomApiDostkOrdrKt10000Response, error) {
	switch side {
	case StockOrderSideBuy:
		return c.PlaceBuyOrder(ctx, req)
	case StockOrderSideSell:
		return c.PlaceSellOrder(ctx, kiwoomspecs.KiwoomApiDostkOrdrKt10001Request{
			DmstStexTp: req.DmstStexTp,
			StkCd:      req.StkCd,
			OrdQty:     req.OrdQty,
			OrdUv:      req.OrdUv,
			TrdeTp:     req.TrdeTp,
			CondUv:     req.CondUv,
		})
	default:
		return nil, broker.ErrInvalidOrderRequest
	}
}

// PlaceBuyOrder places kt10000.
func (c *Client) PlaceBuyOrder(
	ctx context.Context,
	req kiwoomspecs.KiwoomApiDostkOrdrKt10000Request,
) (*kiwoomspecs.KiwoomApiDostkOrdrKt10000Response, error) {
	req.StkCd = normalizeSymbolCode(req.StkCd)
	if req.StkCd == "" || asInt64(req.OrdQty) <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}
	req.DmstStexTp = strings.ToUpper(strings.TrimSpace(req.DmstStexTp))
	if req.DmstStexTp == "" {
		req.DmstStexTp = "KRX"
	}
	req.OrdQty = strings.TrimSpace(req.OrdQty)
	req.OrdUv = strings.TrimSpace(req.OrdUv)
	req.TrdeTp = strings.TrimSpace(req.TrdeTp)
	req.CondUv = strings.TrimSpace(req.CondUv)
	if req.TrdeTp == "" {
		return nil, broker.ErrInvalidOrderRequest
	}

	respObj, err := c.CallDocumentedEndpoint(ctx, "kt10000", PathOrder, &req)
	if err != nil {
		return nil, err
	}
	out := &kiwoomspecs.KiwoomApiDostkOrdrKt10000Response{}
	if err := bindResponseObject(respObj, out); err != nil {
		return nil, err
	}

	if strings.TrimSpace(out.OrdNo) == "" {
		return nil, fmt.Errorf("missing order id in kiwoom response")
	}
	return out, nil
}

// PlaceSellOrder places kt10001.
func (c *Client) PlaceSellOrder(
	ctx context.Context,
	req kiwoomspecs.KiwoomApiDostkOrdrKt10001Request,
) (*kiwoomspecs.KiwoomApiDostkOrdrKt10000Response, error) {
	req.StkCd = normalizeSymbolCode(req.StkCd)
	if req.StkCd == "" || asInt64(req.OrdQty) <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}
	req.DmstStexTp = strings.ToUpper(strings.TrimSpace(req.DmstStexTp))
	if req.DmstStexTp == "" {
		req.DmstStexTp = "KRX"
	}
	req.OrdQty = strings.TrimSpace(req.OrdQty)
	req.OrdUv = strings.TrimSpace(req.OrdUv)
	req.TrdeTp = strings.TrimSpace(req.TrdeTp)
	req.CondUv = strings.TrimSpace(req.CondUv)
	if req.TrdeTp == "" {
		return nil, broker.ErrInvalidOrderRequest
	}

	respObj, err := c.CallDocumentedEndpoint(ctx, "kt10001", PathOrder, &req)
	if err != nil {
		return nil, err
	}
	out := &kiwoomspecs.KiwoomApiDostkOrdrKt10000Response{}
	if err := bindResponseObject(respObj, out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.OrdNo) == "" {
		return nil, fmt.Errorf("missing order id in kiwoom response")
	}
	return out, nil
}

// CancelStockOrder cancels order through kt10003.
func (c *Client) CancelStockOrder(
	ctx context.Context,
	req kiwoomspecs.KiwoomApiDostkOrdrKt10003Request,
) (*kiwoomspecs.KiwoomApiDostkOrdrKt10003Response, error) {
	req.DmstStexTp = strings.ToUpper(strings.TrimSpace(req.DmstStexTp))
	if req.DmstStexTp == "" {
		req.DmstStexTp = "KRX"
	}
	req.StkCd = normalizeSymbolCode(req.StkCd)
	req.OrigOrdNo = strings.TrimSpace(req.OrigOrdNo)
	req.CnclQty = strings.TrimSpace(req.CnclQty)
	if req.StkCd == "" || req.OrigOrdNo == "" || asInt64(req.CnclQty) <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}

	respObj, err := c.CallDocumentedEndpoint(ctx, "kt10003", PathOrder, &req)
	if err != nil {
		return nil, err
	}
	out := &kiwoomspecs.KiwoomApiDostkOrdrKt10003Response{}
	if err := bindResponseObject(respObj, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ModifyStockOrder modifies order through kt10002.
func (c *Client) ModifyStockOrder(
	ctx context.Context,
	req kiwoomspecs.KiwoomApiDostkOrdrKt10002Request,
) (*kiwoomspecs.KiwoomApiDostkOrdrKt10002Response, error) {
	req.DmstStexTp = strings.ToUpper(strings.TrimSpace(req.DmstStexTp))
	if req.DmstStexTp == "" {
		req.DmstStexTp = "KRX"
	}
	req.StkCd = normalizeSymbolCode(req.StkCd)
	req.OrigOrdNo = strings.TrimSpace(req.OrigOrdNo)
	req.MdfyQty = strings.TrimSpace(req.MdfyQty)
	req.MdfyUv = strings.TrimSpace(req.MdfyUv)
	req.MdfyCondUv = strings.TrimSpace(req.MdfyCondUv)
	if req.StkCd == "" || req.OrigOrdNo == "" || asInt64(req.MdfyQty) <= 0 || req.MdfyUv == "" {
		return nil, broker.ErrInvalidOrderRequest
	}

	respObj, err := c.CallDocumentedEndpoint(ctx, "kt10002", PathOrder, &req)
	if err != nil {
		return nil, err
	}
	out := &kiwoomspecs.KiwoomApiDostkOrdrKt10002Response{}
	if err := bindResponseObject(respObj, out); err != nil {
		return nil, err
	}
	return out, nil
}
