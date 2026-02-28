package kis

import (
	"context"
	"fmt"
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
	path := fmt.Sprintf("/uapi/domestic-stock/v1/trading/inquire-balance?CANO=%s&ACNT_PRDT_CD=%s&AFHR_FLPR_YN=N&INQR_DVSN=02&UNPR_DVSN=01&FUND_STTL_ICLD_YN=Y&FNCG_AMT_AUTO_RDPT_YN=Y&OFL_YN=&PRCS_DVSN=01&CTX_AREA_FK100=&CTX_AREA_NK100=",
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

// IsSuccess checks if response is successful
func (r *AccountBalanceResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
