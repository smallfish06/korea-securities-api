package kis

import (
	"context"
	"fmt"
	"net/url"
)

// InquireAccountBalance retrieves comprehensive account balance (stocks + bonds + cash)
// TR_ID: TTTC8434R (실전투자 종합잔고)
func (c *Client) InquireAccountBalance(ctx context.Context, accountNo, accountProductCode string) (*AccountBalanceResponse, error) {
	trID := "TTTC8434R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTC8434R"
	}

	cano := accountNo
	acntPrdtCd := accountProductCode
	if accountProductCode == "" {
		acntPrdtCd = "01"
	}

	// 종합 계좌 조회 (주식+채권+예수금)
	path := fmt.Sprintf("%s?CANO=%s&ACNT_PRDT_CD=%s&AFHR_FLPR_YN=N&INQR_DVSN=02&UNPR_DVSN=01&FUND_STTL_ICLD_YN=Y&FNCG_AMT_AUTO_RDPT_YN=Y&OFL_YN=&PRCS_DVSN=01&CTX_AREA_FK100=&CTX_AREA_NK100=", PathDomesticStockTradingInquireBalance,
		cano, acntPrdtCd)

	var resp AccountBalanceResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire account balance: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}

	return &resp, nil
}

// InquireOverseasBalanceRaw retrieves overseas balance details.
// TR_ID: TTTS3012R (real), VTTS3012R (sandbox)
func (c *Client) InquireOverseasBalanceRaw(ctx context.Context, accountNo, accountProductCode, ovrsExcgCd, trCrcyCd, ctxFK200, ctxNK200 string) (*RawResponse, error) {
	trID := "TTTS3012R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTS3012R"
	}
	if accountProductCode == "" {
		accountProductCode = "01"
	}
	q := url.Values{}
	q.Set("CANO", accountNo)
	q.Set("ACNT_PRDT_CD", accountProductCode)
	q.Set("OVRS_EXCG_CD", ovrsExcgCd)
	q.Set("TR_CRCY_CD", trCrcyCd)
	q.Set("CTX_AREA_FK200", ctxFK200)
	q.Set("CTX_AREA_NK200", ctxNK200)

	return c.getRaw(ctx,
		encodeQuery(PathOverseasStockTradingInquireBalance, q),
		trID,
	)
}

// InquireOverseasPsAmount retrieves overseas orderable amount.
// TR_ID: TTTS3007R (real), VTTS3007R (sandbox)
func (c *Client) InquireOverseasPsAmount(ctx context.Context, accountNo, accountProductCode, ovrsExcgCd, ovrsOrdUnpr, itemCd string) (*RawResponse, error) {
	trID := "TTTS3007R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTS3007R"
	}
	if accountProductCode == "" {
		accountProductCode = "01"
	}
	q := url.Values{}
	q.Set("CANO", accountNo)
	q.Set("ACNT_PRDT_CD", accountProductCode)
	q.Set("OVRS_EXCG_CD", ovrsExcgCd)
	q.Set("OVRS_ORD_UNPR", ovrsOrdUnpr)
	q.Set("ITEM_CD", itemCd)

	return c.getRaw(ctx,
		encodeQuery(PathOverseasStockTradingInquirePsAmount, q),
		trID,
	)
}

// InquirePossibleOrder retrieves domestic orderable amount.
// TR_ID: TTTC8908R (real), VTTC8908R (sandbox)
func (c *Client) InquirePossibleOrder(ctx context.Context, accountNo, accountProductCode, symbol, orderUnitPrice, orderDvsn, cmaEvalIncludedYN, overseasIncludedYN string) (*RawResponse, error) {
	trID := "TTTC8908R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTC8908R"
	}
	if accountProductCode == "" {
		accountProductCode = "01"
	}
	q := url.Values{}
	q.Set("CANO", accountNo)
	q.Set("ACNT_PRDT_CD", accountProductCode)
	q.Set("PDNO", symbol)
	q.Set("ORD_UNPR", orderUnitPrice)
	q.Set("ORD_DVSN", orderDvsn)
	q.Set("CMA_EVLU_AMT_ICLD_YN", cmaEvalIncludedYN)
	q.Set("OVRS_ICLD_YN", overseasIncludedYN)

	return c.getRaw(ctx,
		encodeQuery(PathDomesticStockTradingInquirePsblOrder, q),
		trID,
	)
}

// InquirePeriodTradeProfit retrieves domestic period profit/loss.
// TR_ID: TTTC8715R
func (c *Client) InquirePeriodTradeProfit(ctx context.Context, accountNo, accountProductCode, sortDvsn, startDate, endDate, cblcDvsn, symbol, ctxNK100, ctxFK100 string) (*RawResponse, error) {
	if accountProductCode == "" {
		accountProductCode = "01"
	}
	q := url.Values{}
	q.Set("CANO", accountNo)
	q.Set("ACNT_PRDT_CD", accountProductCode)
	q.Set("SORT_DVSN", sortDvsn)
	q.Set("INQR_STRT_DT", startDate)
	q.Set("INQR_END_DT", endDate)
	q.Set("CBLC_DVSN", cblcDvsn)
	q.Set("PDNO", symbol)
	q.Set("CTX_AREA_NK100", ctxNK100)
	q.Set("CTX_AREA_FK100", ctxFK100)

	return c.getRaw(ctx,
		encodeQuery(PathDomesticStockTradingInquirePeriodTradeProfit, q),
		"TTTC8715R",
	)
}

// IsSuccess checks if response is successful
func (r *AccountBalanceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
