package kis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/smallfish06/kr-broker-api/pkg/broker"
)

// BalanceResponse represents KIS balance response
type BalanceResponse struct {
	RetCode string `json:"rt_cd"`
	MsgCode string `json:"msg_cd"`
	Msg1    string `json:"msg1"`
	Output1 []struct {
		PrdtName     string `json:"prdt_name"`      // 상품명
		HldgQty      string `json:"hldg_qty"`       // 보유수량
		OrdPsblQty   string `json:"ord_psbl_qty"`   // 주문가능수량
		PchsAvgPrce  string `json:"pchs_avg_pric"`  // 매입평균가격
		PchsAmt      string `json:"pchs_amt"`       // 매입금액
		PrprTprt     string `json:"prpr_tprt"`      // 평가손익비율
		EvluAmt      string `json:"evlu_amt"`       // 평가금액
		EvluPflsAmt  string `json:"evlu_pfls_amt"`  // 평가손익금액
		StckLoanUnpr string `json:"stck_loan_unpr"` // 대출단가
		ExpnMgna     string `json:"expn_mgna"`      // 만기보증금액
		FlttRt       string `json:"fltt_rt"`        // 등락율
		BfynFyerDvdn string `json:"bfyn_fyer_dvdn"` // 전년도배당
		StckOprcCurr string `json:"stck_oprc_curr"` // 현재가
		StckSdpr     string `json:"stck_sdpr"`      // 기준가
		StckShrn     string `json:"stck_shrn_iscd"` // 종목코드
	} `json:"output1"`
	Output2 []struct {
		DnCaTotAmt       string `json:"dnca_tot_amt"`          // 예수금총액
		NxdyExccAmt      string `json:"nxdy_excc_amt"`         // D+1추정금액
		PrvsDtCsTotCrAmt string `json:"prvs_dt_cs_tot_cr_amt"` // 전일대비총평가금액
		TotEvluAmt       string `json:"tot_evlu_amt"`          // 총평가금액
		EvluPflsSmtlAmt  string `json:"evlu_pfls_smtl_amt"`    // 평가손익합계금액
		PchsAmtSmtlAmt   string `json:"pchs_amt_smtl_amt"`     // 매입금액합계금액
		EvluAmtSmtlAmt   string `json:"evlu_amt_smtl_amt"`     // 평가금액합계금액
		SllBuyAmtSmtl    string `json:"sll_buy_amt_smtl"`      // 매도매수금액합계
		PnlRt            string `json:"evlu_erng_rt"`          // 손익율
	} `json:"output2"`
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(ctx context.Context, accountID string) (*broker.Balance, error) {
	// KIS API: /uapi/domestic-stock/v1/trading/inquire-balance
	// Query: CANO={accountID앞8자리}&ACNT_PRDT_CD={accountID뒤2자리}&...
	// 예시: 계좌번호가 "12345678-01" 형태라고 가정
	if len(accountID) < 10 {
		return nil, fmt.Errorf("invalid account ID format")
	}

	cano := accountID[:8]
	acntPrdtCd := accountID[9:11]

	path := fmt.Sprintf("/uapi/domestic-stock/v1/trading/inquire-balance?CANO=%s&ACNT_PRDT_CD=%s&AFHR_FLPR_YN=N&OFL_YN=N&INQR_DVSN=01&UNPR_DVSN=01&FUND_STTL_ICLD_YN=N&FNCG_AMT_AUTO_RDPT_YN=N&PRCS_DVSN=00&CTX_AREA_FK100=&CTX_AREA_NK100=",
		cano, acntPrdtCd)

	var resp BalanceResponse
	if err := c.doRequest(ctx, "GET", path, "TTTC8434R", nil, &resp); err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS error: %s - %s", resp.MsgCode, resp.Msg1)
	}

	if len(resp.Output2) == 0 {
		return &broker.Balance{AccountID: accountID}, nil
	}
	out2 := resp.Output2[0]
	cash, _ := strconv.ParseFloat(out2.DnCaTotAmt, 64)
	totalAssets, _ := strconv.ParseFloat(out2.TotEvluAmt, 64)
	profitLoss, _ := strconv.ParseFloat(out2.EvluPflsSmtlAmt, 64)
	profitLossPct, _ := strconv.ParseFloat(out2.PnlRt, 64)

	return &broker.Balance{
		AccountID:        accountID,
		Cash:             cash,
		TotalAssets:      totalAssets,
		BuyingPower:      cash,
		WithdrawableCash: cash,
		ProfitLoss:       profitLoss,
		ProfitLossPct:    profitLossPct,
	}, nil
}

// GetPositions retrieves account positions
func (c *Client) GetPositions(ctx context.Context, accountID string) ([]broker.Position, error) {
	if len(accountID) < 10 {
		return nil, fmt.Errorf("invalid account ID format")
	}

	cano := accountID[:8]
	acntPrdtCd := accountID[9:11]

	path := fmt.Sprintf("/uapi/domestic-stock/v1/trading/inquire-balance?CANO=%s&ACNT_PRDT_CD=%s&AFHR_FLPR_YN=N&OFL_YN=N&INQR_DVSN=01&UNPR_DVSN=01&FUND_STTL_ICLD_YN=N&FNCG_AMT_AUTO_RDPT_YN=N&PRCS_DVSN=00&CTX_AREA_FK100=&CTX_AREA_NK100=",
		cano, acntPrdtCd)

	var resp BalanceResponse
	if err := c.doRequest(ctx, "GET", path, "TTTC8434R", nil, &resp); err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}

	if resp.RetCode != "0" {
		return nil, fmt.Errorf("KIS error: %s - %s", resp.MsgCode, resp.Msg1)
	}

	positions := make([]broker.Position, 0, len(resp.Output1))
	for _, item := range resp.Output1 {
		qty, _ := strconv.ParseInt(item.HldgQty, 10, 64)
		if qty == 0 {
			continue
		}

		avgPrice, _ := strconv.ParseFloat(item.PchsAvgPrce, 64)
		currentPrice, _ := strconv.ParseFloat(item.StckOprcCurr, 64)
		profitLoss, _ := strconv.ParseFloat(item.EvluPflsAmt, 64)
		profitLossPct, _ := strconv.ParseFloat(item.PrprTprt, 64)

		positions = append(positions, broker.Position{
			Symbol:        item.StckShrn,
			Market:        "KRX",
			Quantity:      qty,
			AvgPrice:      avgPrice,
			CurrentPrice:  currentPrice,
			ProfitLoss:    profitLoss,
			ProfitLossPct: profitLossPct,
			WeightPct:     0,
		})
	}

	return positions, nil
}
