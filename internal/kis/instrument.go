package kis

import (
	"context"
	"fmt"
	"net/url"
)

// InquireStockBasicInfo retrieves domestic stock basic info (search-stock-info).
// TR_ID: CTPF1002R
func (c *Client) InquireStockBasicInfo(ctx context.Context, symbol, prdtTypeCode string) (*StockBasicInfoResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if prdtTypeCode == "" {
		prdtTypeCode = "300"
	}

	path := fmt.Sprintf("%s?PRDT_TYPE_CD=%s&PDNO=%s", PathDomesticStockSearchStockInfo,
		url.QueryEscape(prdtTypeCode),
		url.QueryEscape(symbol),
	)

	var resp StockBasicInfoResponse
	if err := c.doRequest(ctx, "GET", path, "CTPF1002R", nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire stock basic info: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireProductBasicInfo retrieves product basic info (search-info).
// TR_ID: CTPF1604R
func (c *Client) InquireProductBasicInfo(ctx context.Context, symbol, prdtTypeCode string) (*ProductBasicInfoResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if prdtTypeCode == "" {
		prdtTypeCode = "300"
	}

	path := fmt.Sprintf("%s?PRDT_TYPE_CD=%s&PDNO=%s", PathDomesticStockSearchInfo,
		url.QueryEscape(prdtTypeCode),
		url.QueryEscape(symbol),
	)

	var resp ProductBasicInfoResponse
	if err := c.doRequest(ctx, "GET", path, "CTPF1604R", nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire product basic info: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireOverseasProductBasicInfo retrieves overseas product basic info.
// TR_ID: CTPF1702R
func (c *Client) InquireOverseasProductBasicInfo(ctx context.Context, symbol, prdtTypeCode string) (*OverseasProductBasicInfoResponse, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if prdtTypeCode == "" {
		prdtTypeCode = "512"
	}

	path := fmt.Sprintf("%s?PRDT_TYPE_CD=%s&PDNO=%s", PathOverseasPriceSearchInfo,
		url.QueryEscape(prdtTypeCode),
		url.QueryEscape(symbol),
	)

	var resp OverseasProductBasicInfoResponse
	if err := c.doRequest(ctx, "GET", path, "CTPF1702R", nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire overseas product basic info: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

func (r *StockBasicInfoResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *ProductBasicInfoResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *OverseasProductBasicInfoResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
