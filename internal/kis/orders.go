package kis

import (
	"context"
	"fmt"
	"time"

	"github.com/smallfish06/krsec/pkg/broker"
)

// OrderResponse represents KIS order response
type OrderResponse struct {
	RetCode string `json:"rt_cd"`
	MsgCode string `json:"msg_cd"`
	Msg1    string `json:"msg1"`
	Output  struct {
		KrxFwdOrdOrgno string `json:"KRX_FWDG_ORD_ORGNO"` // 주문조직번호
		OrdNo          string `json:"ODNO"`               // 주문번호
		OrdTmd         string `json:"ORD_TMD"`            // 주문시각
	} `json:"output"`
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, req broker.OrderRequest) (*broker.OrderResult, error) {
	if len(req.AccountID) < 10 {
		return nil, fmt.Errorf("invalid account ID format")
	}

	cano := req.AccountID[:8]
	acntPrdtCd := req.AccountID[9:11]

	// 매수/매도 구분
	var trID string
	var ordDvsn string

	if req.Type == broker.OrderTypeMarket {
		ordDvsn = "01" // 시장가
	} else {
		ordDvsn = "00" // 지정가
	}

	if req.Side == broker.OrderSideBuy {
		trID = "TTTC0802U" // 매수
	} else {
		trID = "TTTC0801U" // 매도
	}

	reqBody := map[string]interface{}{
		"CANO":         cano,
		"ACNT_PRDT_CD": acntPrdtCd,
		"PDNO":         req.Symbol,
		"ORD_DVSN":     ordDvsn,
		"ORD_QTY":      fmt.Sprintf("%d", req.Quantity),
		"ORD_UNPR":     fmt.Sprintf("%.0f", req.Price),
	}

	var resp OrderResponse
	if err := c.doRequest(ctx, "POST", PathDomesticStockTradingOrderCash, trID, reqBody, &resp); err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	if resp.RetCode != "0" {
		return &broker.OrderResult{
			OrderID:        "",
			Status:         broker.OrderStatusRejected,
			RejectedReason: resp.Msg1,
			Message:        fmt.Sprintf("%s - %s", resp.MsgCode, resp.Msg1),
			Timestamp:      time.Now(),
		}, nil
	}

	return &broker.OrderResult{
		OrderID:      resp.Output.OrdNo,
		Status:       broker.OrderStatusPending,
		RemainingQty: req.Quantity,
		Message:      resp.Msg1,
		Timestamp:    time.Now(),
	}, nil
}

// CancelOrder cancels an order
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	return fmt.Errorf("%w: use OrderRvseCncl with account/order context", broker.ErrInvalidOrderRequest)
}

// ModifyOrder modifies an existing order
func (c *Client) ModifyOrder(ctx context.Context, orderID string, req broker.ModifyOrderRequest) (*broker.OrderResult, error) {
	return nil, fmt.Errorf("%w: use OrderRvseCncl with account/order context", broker.ErrInvalidOrderRequest)
}
