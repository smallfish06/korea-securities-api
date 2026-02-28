package kis

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// InquireDailyCcld retrieves domestic daily order/filled status.
// TR_ID: TTTC0081R (real), VTTC0081R (sandbox)
func (c *Client) InquireDailyCcld(
	ctx context.Context,
	accountNo, accountProductCode, startDate, endDate, orderOrgNo, orderNo, exchangeID string,
) (*DomesticDailyCcldResponse, error) {
	trID := "TTTC0081R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTC0081R"
	}

	if accountProductCode == "" {
		accountProductCode = "01"
	}
	if startDate == "" || endDate == "" {
		today := time.Now().Format("20060102")
		if startDate == "" {
			startDate = today
		}
		if endDate == "" {
			endDate = today
		}
	}
	if exchangeID == "" {
		exchangeID = "ALL"
	}

	q := url.Values{}
	q.Set("CANO", accountNo)
	q.Set("ACNT_PRDT_CD", accountProductCode)
	q.Set("INQR_STRT_DT", startDate)
	q.Set("INQR_END_DT", endDate)
	q.Set("SLL_BUY_DVSN_CD", "00")
	q.Set("INQR_DVSN", "00")
	q.Set("PDNO", "")
	q.Set("CCLD_DVSN", "00")
	q.Set("ORD_GNO_BRNO", orderOrgNo)
	q.Set("ODNO", orderNo)
	q.Set("INQR_DVSN_3", "00")
	q.Set("INQR_DVSN_1", "")
	q.Set("EXCG_ID_DVSN_CD", exchangeID)
	q.Set("CTX_AREA_FK100", "")
	q.Set("CTX_AREA_NK100", "")

	path := PathDomesticStockTradingInquireDailyCcld + "?" + q.Encode()

	var resp DomesticDailyCcldResponse
	if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
		return nil, fmt.Errorf("inquire daily ccld: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
	}
	return &resp, nil
}

// InquireOverseasCcnl retrieves overseas order/fill status.
// TR_ID: TTTS3035R (real), VTTS3035R (sandbox)
func (c *Client) InquireOverseasCcnl(
	ctx context.Context,
	accountNo, accountProductCode, startDate, endDate, exchangeCode string,
) (*OverseasCcnlResponse, error) {
	trID := "TTTS3035R"
	if c.baseURL == BaseURLSandbox {
		trID = "VTTS3035R"
	}

	if accountProductCode == "" {
		accountProductCode = "01"
	}
	if startDate == "" || endDate == "" {
		today := time.Now().Format("20060102")
		if startDate == "" {
			startDate = today
		}
		if endDate == "" {
			endDate = today
		}
	}
	if exchangeCode == "" {
		exchangeCode = "%"
	}

	ctxFK := ""
	ctxNK := ""
	all := make([]OverseasCcnlItem, 0)
	seenCursors := make(map[string]struct{})
	const maxPages = 200

	for i := 0; i < maxPages; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		q := url.Values{}
		q.Set("CANO", accountNo)
		q.Set("ACNT_PRDT_CD", accountProductCode)
		q.Set("PDNO", "%")
		q.Set("ORD_STRT_DT", startDate)
		q.Set("ORD_END_DT", endDate)
		q.Set("SLL_BUY_DVSN", "00")
		q.Set("CCLD_NCCS_DVSN", "00")
		q.Set("OVRS_EXCG_CD", exchangeCode)
		q.Set("SORT_SQN", "DS")
		q.Set("ORD_DT", "")
		q.Set("ORD_GNO_BRNO", "")
		q.Set("ODNO", "")
		q.Set("CTX_AREA_NK200", ctxNK)
		q.Set("CTX_AREA_FK200", ctxFK)

		path := PathOverseasStockTradingInquireCcnl + "?" + q.Encode()

		var resp OverseasCcnlResponse
		if err := c.doRequest(ctx, "GET", path, trID, nil, &resp); err != nil {
			return nil, fmt.Errorf("inquire overseas ccnl: %w", err)
		}
		if !resp.IsSuccess() {
			return nil, fmt.Errorf("KIS API error: %s (%s)", resp.Msg1, resp.MsgCD)
		}

		all = append(all, resp.Output...)

		nextFK := strings.TrimSpace(resp.CtxAreaFK200)
		nextNK := strings.TrimSpace(resp.CtxAreaNK200)
		if nextNK == "" || (nextFK == ctxFK && nextNK == ctxNK) {
			break
		}
		cursorKey := nextFK + "|" + nextNK
		if _, exists := seenCursors[cursorKey]; exists {
			break
		}
		seenCursors[cursorKey] = struct{}{}
		ctxFK = nextFK
		ctxNK = nextNK
	}

	if len(all) > 0 {
		return &OverseasCcnlResponse{
			RtCD:   "0",
			MsgCD:  "00000",
			Msg1:   "정상처리 되었습니다.",
			Output: all,
		}, nil
	}

	return &OverseasCcnlResponse{
		RtCD:   "0",
		MsgCD:  "00000",
		Msg1:   "정상처리 되었습니다.",
		Output: []OverseasCcnlItem{},
	}, nil
}

func (r *DomesticDailyCcldResponse) IsSuccess() bool {
	return r.RtCD == "0"
}

func (r *OverseasCcnlResponse) IsSuccess() bool {
	return r.RtCD == "0"
}
