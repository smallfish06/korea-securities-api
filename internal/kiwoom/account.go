package kiwoom

import (
	"context"
	"strings"
)

// GetAccountBalance fetches kt00005.
func (c *Client) GetAccountBalance(ctx context.Context, exchange string) (*AccountBalance, error) {
	exchange = strings.ToUpper(strings.TrimSpace(exchange))
	if exchange == "" {
		exchange = "KRX"
	}

	res, err := c.call(ctx, "kt00005", map[string]interface{}{
		"dmst_stex_tp": exchange,
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	return &AccountBalance{
		Deposit:              asFloat64(res.Body["entr"]),
		DepositD1:            asFloat64(res.Body["entr_d1"]),
		DepositD2:            asFloat64(res.Body["entr_d2"]),
		OrderableAmount:      asFloat64(res.Body["ord_alowa"]),
		WithdrawableAmount:   asFloat64(res.Body["wthd_alowa"]),
		UnsettledStockAmount: asFloat64(res.Body["uncl_stk_amt"]),
		StockBuyTotalAmount:  asFloat64(res.Body["stk_buy_tot_amt"]),
		EvaluationTotal:      asFloat64(res.Body["evlt_amt_tot"]),
		TotalProfitLoss:      asFloat64(res.Body["tot_pl_tot"]),
		TotalProfitLossRate:  asFloat64(res.Body["tot_pl_rt"]),
		PresumedAssetAmount:  asFloat64(res.Body["prsm_dpst_aset_amt"]),
		CreditLoanTotal:      asFloat64(res.Body["crd_loan_tot"]),
		ReturnMsg:            asString(res.Body["return_msg"]),
		ReturnCode:           parseReturnCode(res.Body["return_code"]),
	}, nil
}

// GetAccountPositions fetches kt00018.
func (c *Client) GetAccountPositions(ctx context.Context, queryType, exchange string) ([]AccountPosition, error) {
	queryType = strings.TrimSpace(queryType)
	if queryType == "" {
		queryType = "0"
	}
	exchange = strings.ToUpper(strings.TrimSpace(exchange))
	if exchange == "" {
		exchange = "KRX"
	}

	res, err := c.call(ctx, "kt00018", map[string]interface{}{
		"qry_tp":       queryType,
		"dmst_stex_tp": exchange,
	}, callOptions{})
	if err != nil {
		return nil, err
	}

	rows := firstObjectArray(res.Body, "acnt_evlt_remn_indv_tot")
	if len(rows) == 0 {
		return []AccountPosition{}, nil
	}

	positions := make([]AccountPosition, 0, len(rows))
	for _, row := range rows {
		positions = append(positions, AccountPosition{
			StockCode:        normalizeSymbolCode(asString(row["stk_cd"])),
			StockName:        asString(row["stk_nm"]),
			RemainingQty:     asInt64(row["rmnd_qty"]),
			TradableQty:      asInt64(row["trde_able_qty"]),
			TodayBuyQty:      asInt64(row["tdy_buyq"]),
			TodaySellQty:     asInt64(row["tdy_sellq"]),
			PurchasePrice:    asFloat64(row["pur_pric"]),
			CurrentPrice:     asFloat64(row["cur_prc"]),
			PurchaseAmount:   asFloat64(row["pur_amt"]),
			EvaluationAmount: asFloat64(row["evlt_amt"]),
			EvaluationProfit: asFloat64(row["evltv_prft"]),
			ProfitRate:       asFloat64(row["prft_rt"]),
			WeightRate:       asFloat64(row["poss_rt"]),
			CreditLoanDate:   asString(row["crd_loan_dt"]),
		})
	}
	return positions, nil
}

// GetUnsettledOrders fetches ka10075.
func (c *Client) GetUnsettledOrders(ctx context.Context, symbol string) ([]UnsettledOrder, error) {
	body := map[string]interface{}{
		"all_stk_tp": "0",
		"trde_tp":    "0",
		"stex_tp":    "0",
	}
	symbol = normalizeSymbolCode(symbol)
	if symbol != "" {
		body["stk_cd"] = symbol
		body["all_stk_tp"] = "1"
	}

	res, err := c.call(ctx, "ka10075", body, callOptions{})
	if err != nil {
		return nil, err
	}

	rows := firstObjectArray(res.Body, "oso")
	if len(rows) == 0 {
		return []UnsettledOrder{}, nil
	}

	orders := make([]UnsettledOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, UnsettledOrder{
			OrderNumber:    asString(row["ord_no"]),
			StockCode:      normalizeSymbolCode(asString(row["stk_cd"])),
			OrderStatus:    asString(row["ord_stt"]),
			OrderQty:       asInt64(row["ord_qty"]),
			UnsettledQty:   asInt64(row["oso_qty"]),
			OrderPrice:     asFloat64(row["ord_pric"]),
			ConcludedPrice: asFloat64(row["cntr_pric"]),
			OrderSideText:  asString(row["io_tp_nm"]),
			ExchangeCode:   asString(row["stex_tp"]),
			ExchangeText:   asString(row["stex_tp_txt"]),
			ReturnMsg:      asString(res.Body["return_msg"]),
			ReturnCode:     parseReturnCode(res.Body["return_code"]),
		})
	}
	return orders, nil
}

// GetOrderExecutions fetches ka10076.
func (c *Client) GetOrderExecutions(ctx context.Context, symbol string) ([]OrderExecution, error) {
	body := map[string]interface{}{
		"qry_tp":  "0",
		"sell_tp": "0",
		"stex_tp": "0",
	}
	symbol = normalizeSymbolCode(symbol)
	if symbol != "" {
		body["stk_cd"] = symbol
	}

	res, err := c.call(ctx, "ka10076", body, callOptions{})
	if err != nil {
		return nil, err
	}

	rows := firstObjectArray(res.Body, "cntr")
	if len(rows) == 0 {
		return []OrderExecution{}, nil
	}

	executions := make([]OrderExecution, 0, len(rows))
	for _, row := range rows {
		executions = append(executions, OrderExecution{
			OrderNumber:    asString(row["ord_no"]),
			StockCode:      normalizeSymbolCode(asString(row["stk_cd"])),
			OrderSideText:  asString(row["io_tp_nm"]),
			ExecutionPrice: asFloat64(row["cntr_pric"]),
			ExecutionQty:   asInt64(row["cntr_qty"]),
			OrderTime:      asString(row["ord_tm"]),
			OrderStatus:    asString(row["ord_stt"]),
			ExchangeCode:   asString(row["stex_tp"]),
			ExchangeText:   asString(row["stex_tp_txt"]),
		})
	}
	return executions, nil
}
