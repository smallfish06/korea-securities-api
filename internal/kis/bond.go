package kis

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// InquireBondPrice retrieves current bond price
// TR_ID: FHKST03010100 (국내채권 시세조회)
func (c *Client) InquireBondPrice(ctx context.Context, isinCode string) (*BondPriceResponse, error) {
	trID := "FHKST03010100"
	if c.baseURL == BaseURLSandbox {
		trID = "VHKST03010100" // 모의투자용 (추정)
	}

	today := time.Now().Format("20060102")
	path := fmt.Sprintf(
		"%s?FID_COND_MRKT_DIV_CODE=J&FID_INPUT_ISCD=%s&FID_INPUT_DATE_1=%s&FID_INPUT_DATE_2=%s&FID_PERIOD_DIV_CODE=D&FID_ORG_ADJ_PRC=0",
		PathDomesticBondInquirePrice,
		isinCode, today, today,
	)

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
		path := fmt.Sprintf("%s?CANO=%s&ACNT_PRDT_CD=%s&INQR_CNDT=00&PDNO=&BUY_DT=&CTX_AREA_FK200=%s&CTX_AREA_NK200=%s", PathDomesticBondInquireBalance,
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

// InquireBondDaily retrieves daily bond price data.
// KIS domestic-bond endpoint currently uses the same service(TR_ID) and required
// query schema as inquire-price for historical 조회.
// TR_ID: FHKST03010100
func (c *Client) InquireBondDaily(ctx context.Context, isinCode string, startDate, endDate string) (*StockDailyPriceResponse, error) {
	trID := "FHKST03010100"
	if c.baseURL == BaseURLSandbox {
		trID = "VHKST03010100"
	}
	if strings.TrimSpace(startDate) == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("20060102")
	}
	if strings.TrimSpace(endDate) == "" {
		endDate = time.Now().Format("20060102")
	}

	path := fmt.Sprintf(
		"%s?FID_COND_MRKT_DIV_CODE=J&FID_INPUT_ISCD=%s&FID_INPUT_DATE_1=%s&FID_INPUT_DATE_2=%s&FID_PERIOD_DIV_CODE=D&FID_ORG_ADJ_PRC=0",
		PathDomesticBondInquirePrice,
		isinCode, startDate, endDate,
	)

	var resp struct {
		RtCD    string                  `json:"rt_cd"`
		MsgCD   string                  `json:"msg_cd"`
		Msg1    string                  `json:"msg1"`
		Output1 map[string]interface{}  `json:"output1"`
		Output2 []StockDailyPriceOutput `json:"output2"`
	}
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire bond daily: %w", err)
	}

	if resp.RtCD != "0" {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &StockDailyPriceResponse{
		RtCD:    resp.RtCD,
		MsgCD:   resp.MsgCD,
		Msg1:    resp.Msg1,
		Output:  resp.Output2,
		Output1: resp.Output2,
	}, nil
}

// IsSuccess checks if response is successful
func (r *BondPriceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *BondBalanceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
