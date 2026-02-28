package kis

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// InquireBondPrice retrieves current bond price
// TR_ID: FHKST03010100 (국내채권 시세조회)
func (c *Client) InquireBondPrice(ctx context.Context, isinCode string) (*BondPriceResponse, error) {
	trID := "FHKST03010100"
	if c.baseURL == BaseURLSandbox {
		trID = "VHKST03010100" // 모의투자용 (추정)
	}

	path := fmt.Sprintf("/uapi/domestic-bond/v1/quotations/inquire-price?fid_input_iscd=%s", isinCode)

	var resp BondPriceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire bond price: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireBondBalance retrieves bond account balance with pagination
// TR_ID: CTSC8407R (장내채권 잔고조회)
func (c *Client) InquireBondBalance(ctx context.Context, accountNo, accountProductCode string) (*BondBalanceResponse, error) {
	trID := "CTSC8407R"

	cano := accountNo
	acntPrdtCd := accountProductCode
	if accountProductCode == "" {
		acntPrdtCd = "01"
	}

	var allOutput []BondBalanceOutput
	var allOutput1 []BondBalanceOutput
	seen := make(map[string]bool) // deduplicate by pdno+buy_dt+buy_sqno
	ctxFK := ""
	ctxNK := ""
	maxPages := 10

	for page := 0; page < maxPages; page++ {
		path := fmt.Sprintf("/uapi/domestic-bond/v1/trading/inquire-balance?CANO=%s&ACNT_PRDT_CD=%s&INQR_CNDT=00&PDNO=&BUY_DT=&CTX_AREA_FK200=%s&CTX_AREA_NK200=%s",
			cano, acntPrdtCd, url.QueryEscape(strings.TrimSpace(ctxFK)), url.QueryEscape(strings.TrimSpace(ctxNK)))

		var resp BondBalanceResponse
		if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
			return nil, fmt.Errorf("inquire bond balance: %w", err)
		}

		if !resp.IsSuccess() {
			return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
		}

		// Merge results (deduplicated)
		for _, item := range resp.Output {
			key := item.PdNo + "|" + item.BuyDt + "|" + item.BuySqno
			if !seen[key] {
				seen[key] = true
				allOutput = append(allOutput, item)
			}
		}
		for _, item := range resp.Output1 {
			key := item.PdNo + "|" + item.BuyDt + "|" + item.BuySqno
			if !seen[key] {
				seen[key] = true
				allOutput1 = append(allOutput1, item)
			}
		}

		// Check for next page - KIS signals continuation via msg1 containing "조회가 계속됩니다"
		if !strings.Contains(resp.Msg1, "조회가 계속") || strings.TrimSpace(resp.CtxAreaNK200) == "" {
			break
		}

		newFK := strings.TrimSpace(resp.CtxAreaFK200)
		newNK := strings.TrimSpace(resp.CtxAreaNK200)

		// Guard against infinite loop with same pagination keys
		if newFK == ctxFK && newNK == ctxNK {
			break
		}

		ctxFK = newFK
		ctxNK = newNK
	}

	return &BondBalanceResponse{
		RtCD:         "0",
		MsgCD:        "00000",
		Msg1:         "정상처리 되었습니다.",
		Output:       allOutput,
		Output1:      allOutput1,
		CtxAreaFK200: "",
		CtxAreaNK200: "",
	}, nil
}

// InquireBondDaily retrieves daily bond price data
// TR_ID: FHKST03010300 (채권 일별 시세 - 추정)
func (c *Client) InquireBondDaily(ctx context.Context, isinCode string, startDate, endDate string) (*StockDailyPriceResponse, error) {
	// Note: 채권 일봉 조회 API는 실제 문서 확인 필요
	trID := "FHKST03010300"
	if c.baseURL == BaseURLSandbox {
		trID = "VHKST03010300"
	}

	path := fmt.Sprintf("/uapi/domestic-bond/v1/quotations/inquire-daily-price?fid_input_iscd=%s&fid_period_div_code=D",
		isinCode)

	var resp StockDailyPriceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire bond daily: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// IsSuccess checks if response is successful
func (r *BondPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *BondBalanceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
