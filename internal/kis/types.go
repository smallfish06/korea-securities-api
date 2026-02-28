package kis

import "time"

// ====================
// Stock Quote
// ====================

// StockPriceResponse represents KIS stock price inquiry response
type StockPriceResponse struct {
	RtCD    string           `json:"rt_cd"`   // 성공 실패 여부 (0: 성공)
	MsgCD   string           `json:"msg_cd"`  // 응답코드
	Msg1    string           `json:"msg1"`    // 응답메세지
	Output  StockPriceOutput `json:"output"`  // 응답상세
	Output1 StockPriceOutput `json:"output1"` // 응답상세 (일부 API는 output1 사용)
}

// StockPriceOutput represents stock price data
type StockPriceOutput struct {
	StckPrpr     string `json:"stck_prpr"`      // 주식 현재가
	PrdyVrss     string `json:"prdy_vrss"`      // 전일 대비
	PrdyVrssSign string `json:"prdy_vrss_sign"` // 전일 대비 부호
	PrdyCtrt     string `json:"prdy_ctrt"`      // 전일 대비율
	AcmlVol      string `json:"acml_vol"`       // 누적 거래량
	AcmlTrPbmn   string `json:"acml_tr_pbmn"`   // 누적 거래 대금
	StckOprc     string `json:"stck_oprc"`      // 주식 시가
	StckHgpr     string `json:"stck_hgpr"`      // 주식 최고가
	StckLwpr     string `json:"stck_lwpr"`      // 주식 최저가
	StckMxpr     string `json:"stck_mxpr"`      // 주식 상한가
	StckLlam     string `json:"stck_llam"`      // 주식 하한가
	PrdyVol      string `json:"prdy_vol"`       // 전일 거래량
}

// ====================
// Stock Daily Price (OHLCV)
// ====================

// StockDailyPriceResponse represents KIS daily price inquiry response
type StockDailyPriceResponse struct {
	RtCD   string                  `json:"rt_cd"`
	MsgCD  string                  `json:"msg_cd"`
	Msg1   string                  `json:"msg1"`
	Output []StockDailyPriceOutput `json:"output2"` // 일봉은 output2
}

// StockDailyPriceOutput represents daily OHLCV data
type StockDailyPriceOutput struct {
	StckBsopDate string `json:"stck_bsop_date"` // 주식 영업 일자
	StckClpr     string `json:"stck_clpr"`      // 주식 종가
	StckOprc     string `json:"stck_oprc"`      // 주식 시가
	StckHgpr     string `json:"stck_hgpr"`      // 주식 최고가
	StckLwpr     string `json:"stck_lwpr"`      // 주식 최저가
	AcmlVol      string `json:"acml_vol"`       // 누적 거래량
	AcmlTrPbmn   string `json:"acml_tr_pbmn"`   // 누적 거래 대금
	PrdyVrss     string `json:"prdy_vrss"`      // 전일 대비
	PrdyVrssSign string `json:"prdy_vrss_sign"` // 전일 대비 부호
}

// ====================
// Stock Balance
// ====================

// StockBalanceResponse represents KIS stock balance inquiry response
type StockBalanceResponse struct {
	RtCD    string                `json:"rt_cd"`
	MsgCD   string                `json:"msg_cd"`
	Msg1    string                `json:"msg1"`
	Output1 []StockBalanceOutput  `json:"output1"` // 보유 종목 리스트
	Output2 []StockBalanceSummary `json:"output2"` // 계좌 요약
}

// StockBalanceOutput represents a single stock position
type StockBalanceOutput struct {
	PdNo         string `json:"pdno"`           // 종목코드
	PrdtName     string `json:"prdt_name"`      // 상품명
	HldgQty      string `json:"hldg_qty"`       // 보유수량
	OrdPsblQty   string `json:"ord_psbl_qty"`   // 주문가능수량
	PchmAvgPric  string `json:"pchs_avg_pric"`  // 매입평균가격
	Prpr         string `json:"prpr"`           // 현재가
	PrprTprt     string `json:"prpr_tprt"`      // 현재가 구분
	EvluPflsAmt  string `json:"evlu_pfls_amt"`  // 평가손익금액
	EvluPflsRt   string `json:"evlu_pfls_rt"`   // 평가손익율
	EvluAmt      string `json:"evlu_amt"`       // 평가금액
	PchmAmt      string `json:"pchm_amt"`       // 매입금액
	HldgQtyRatio string `json:"hldg_qty_ratio"` // 보유비중
}

// StockBalanceSummary represents account balance summary
type StockBalanceSummary struct {
	DnpspCblAmt  string `json:"dnca_tot_amt"`       // 예수금총액
	NxdyExccAmt  string `json:"nxdy_excc_amt"`      // 익일정산금액
	PrvsRcdlExcc string `json:"prvs_rcdl_excc"`     // 가수도정산금액
	CmaEvluAmt   string `json:"cma_evlu_amt"`       // CMA평가금액
	BfmBnysTotAm string `json:"bfdy_buy_amt"`       // 전일매수금액
	ThdtBuyAmt   string `json:"thdt_buy_amt"`       // 금일매수금액
	NxdySellAmt  string `json:"nxdy_auto_rdpt_amt"` // 익일자동상환금액
	DpsplTotAmt  string `json:"tot_evlu_amt"`       // 총평가금액
	PchsAmtSmtl  string `json:"pchs_amt_smtl_amt"`  // 매입금액합계금액
	EvluAmtSmtl  string `json:"evlu_amt_smtl_amt"`  // 평가금액합계금액
	EvluPflsSm   string `json:"evlu_pfls_smtl_amt"` // 평가손익합계금액
	TotEvluPfls  string `json:"tot_evlu_pfls_amt"`  // 총평가손익금액
	TotStlnSlCa  string `json:"tot_stln_slng_chgs"` // 총대출금액
	EvluErngRt   string `json:"evlu_erng_rt"`       // 평가수익율
}

// ====================
// Bond Quote
// ====================

// BondPriceResponse represents KIS bond price inquiry response
type BondPriceResponse struct {
	RtCD   string          `json:"rt_cd"`
	MsgCD  string          `json:"msg_cd"`
	Msg1   string          `json:"msg1"`
	Output BondPriceOutput `json:"output"`
}

// BondPriceOutput represents bond price data
type BondPriceOutput struct {
	IsinCd       string `json:"isin_cd"`        // ISIN코드
	IsinCdNm     string `json:"isin_cd_nm"`     // ISIN코드명
	BondPblcPric string `json:"bond_pblc_pric"` // 채권공모가격
	BondStprPric string `json:"bond_stpr_pric"` // 채권기준가격
	BondClsgPric string `json:"bond_clsg_pric"` // 채권종가
	BondPrdyVrss string `json:"bond_prdy_vrss"` // 채권전일대비
	BondOprcPric string `json:"bond_oprc_pric"` // 채권시가
	BondHgprPric string `json:"bond_hgpr_pric"` // 채권고가
	BondLwprPric string `json:"bond_lwpr_pric"` // 채권저가
	AcmlVol      string `json:"acml_vol"`       // 누적거래량
	AcmlTrPbmn   string `json:"acml_tr_pbmn"`   // 누적거래대금
}

// ====================
// Bond Balance
// ====================

// BondBalanceResponse represents KIS bond balance inquiry response
type BondBalanceResponse struct {
	RtCD         string              `json:"rt_cd"`
	MsgCD        string              `json:"msg_cd"`
	Msg1         string              `json:"msg1"`
	Output       []BondBalanceOutput `json:"output"`  // 보유 채권 리스트
	Output1      []BondBalanceOutput `json:"output1"` // 보유 채권 리스트 (fallback)
	CtxAreaFK200 string              `json:"ctx_area_fk200"`
	CtxAreaNK200 string              `json:"ctx_area_nk200"`
}

// BondBalanceOutput represents a single bond position
type BondBalanceOutput struct {
	PdNo       string `json:"pdno"`         // 종목코드
	PrdtName   string `json:"prdt_name"`    // 종목명
	CblcQty    string `json:"cblc_qty"`     // 잔고수량
	BuyUnpr    string `json:"buy_unpr"`     // 매입단가
	BuyAmt     string `json:"buy_amt"`      // 매입금액
	BuyDt      string `json:"buy_dt"`       // 매수일자
	BuySqno    string `json:"buy_sqno"`     // 매수순번
	AgrxQty    string `json:"agrx_qty"`     // 총세수량
	SprxQty    string `json:"sprx_qty"`     // 분리과세수량
	Exdt       string `json:"exdt"`         // 만기일
	BuyErngRt  string `json:"buy_erng_rt"`  // 매수수익률
	OrdPsblQty string `json:"ord_psbl_qty"` // 주문가능수량
}

// BondBalanceSummary represents bond account summary
type BondBalanceSummary struct {
	TotEvluAmt     string `json:"tot_evlu_amt"`       // 총평가금액
	PchsAmtSmtlAmt string `json:"pchs_amt_smtl_amt"`  // 매입금액합계
	EvluPflsSmtl   string `json:"evlu_pfls_smtl_amt"` // 평가손익합계
	EvluErngRt     string `json:"evlu_erng_rt"`       // 평가수익율
}

// ====================
// Account Balance (종합)
// ====================

// AccountBalanceResponse represents comprehensive account balance
type AccountBalanceResponse struct {
	RtCD   string               `json:"rt_cd"`
	MsgCD  string               `json:"msg_cd"`
	Msg1   string               `json:"msg1"`
	Output AccountBalanceOutput `json:"output"`
}

// AccountBalanceOutput represents comprehensive account data
type AccountBalanceOutput struct {
	DnpspCblAmt string `json:"dnca_tot_amt"`       // 예수금총액
	TotEvluAmt  string `json:"tot_evlu_amt"`       // 총평가금액
	PchsAmtSmtl string `json:"pchs_amt_smtl_amt"`  // 매입금액합계
	EvluPflsSm  string `json:"evlu_pfls_smtl_amt"` // 평가손익합계
	EvluErngRt  string `json:"evlu_erng_rt"`       // 평가수익율
	SsetTotAmt  string `json:"sset_tot_amt"`       // 자산총액
}

// ====================
// Overseas Stock
// ====================

// OverseasPriceResponse represents overseas stock price response
type OverseasPriceResponse struct {
	RtCD   string              `json:"rt_cd"`
	MsgCD  string              `json:"msg_cd"`
	Msg1   string              `json:"msg1"`
	Output OverseasPriceOutput `json:"output"`
}

// OverseasPriceOutput represents overseas stock price data
type OverseasPriceOutput struct {
	RsymStr      string `json:"rsym"`           // 실시간종목코드
	SymbDesc     string `json:"symb_desc"`      // 종목명
	Last         string `json:"last"`           // 현재가
	Open         string `json:"open"`           // 시가
	High         string `json:"high"`           // 고가
	Low          string `json:"low"`            // 저가
	PrdyVrss     string `json:"prdy_vrss"`      // 전일대비
	PrdyVrssSign string `json:"prdy_vrss_sign"` // 전일대비부호
	AccrTrVol    string `json:"t_xvol"`         // 누적거래량
}

// ====================
// Instrument Info
// ====================

// StockBasicInfoResponse represents domestic stock basic info response (search-stock-info).
type StockBasicInfoResponse struct {
	RtCD   string               `json:"rt_cd"`
	MsgCD  string               `json:"msg_cd"`
	Msg1   string               `json:"msg1"`
	Output StockBasicInfoOutput `json:"output"`
}

// StockBasicInfoOutput represents domestic stock basic info payload.
type StockBasicInfoOutput struct {
	PdNo                 string `json:"pdno"`
	PrdtTypeCD           string `json:"prdt_type_cd"`
	MketIDCD             string `json:"mket_id_cd"`
	SctyGrpIDCD          string `json:"scty_grp_id_cd"`
	ExcgDvsnCD           string `json:"excg_dvsn_cd"`
	PrdtName             string `json:"prdt_name"`
	PrdtName120          string `json:"prdt_name120"`
	PrdtAbrvName         string `json:"prdt_abrv_name"`
	PrdtEngName          string `json:"prdt_eng_name"`
	PrdtEngName120       string `json:"prdt_eng_name120"`
	PrdtEngAbrvName      string `json:"prdt_eng_abrv_name"`
	StdPdNo              string `json:"std_pdno"`
	PrdtClsfCD           string `json:"prdt_clsf_cd"`
	PrdtClsfName         string `json:"prdt_clsf_name"`
	StdIdstClsfCDName    string `json:"std_idst_clsf_cd_name"`
	IdxBztpLclsCDName    string `json:"idx_bztp_lcls_cd_name"`
	IdxBztpMclsCDName    string `json:"idx_bztp_mcls_cd_name"`
	IdxBztpSclsCDName    string `json:"idx_bztp_scls_cd_name"`
	TrStopYN             string `json:"tr_stop_yn"`
	LstgAbolDt           string `json:"lstg_abol_dt"`
	SctsMketLstgDt       string `json:"scts_mket_lstg_dt"`
	KosdaqMketLstgDt     string `json:"kosdaq_mket_lstg_dt"`
	CpttTradTrPsblYN     string `json:"cptt_trad_tr_psbl_yn"`
	NxtTrStopYN          string `json:"nxt_tr_stop_yn"`
	StdPdNoShort         string `json:"shtn_pdno"`
	PrdtSaleStatCD       string `json:"prdt_sale_stat_cd"`
	PrdtRiskGradCD       string `json:"prdt_risk_grad_cd"`
	StdIdstClsfCD        string `json:"std_idst_clsf_cd"`
	LstgStqt             string `json:"lstg_stqt"`
	ThdtClpr             string `json:"thdt_clpr"`
	BfdyClpr             string `json:"bfdy_clpr"`
	FrbdMketLstgDt       string `json:"frbd_mket_lstg_dt"`
	FrbdMketLstgAbolDt   string `json:"frbd_mket_lstg_abol_dt"`
	SctsMketLstgAbolDt   string `json:"scts_mket_lstg_abol_dt"`
	KosdaqMketLstgAbolDt string `json:"kosdaq_mket_lstg_abol_dt"`
}

// ProductBasicInfoResponse represents product basic info response (search-info).
type ProductBasicInfoResponse struct {
	RtCD   string                 `json:"rt_cd"`
	MsgCD  string                 `json:"msg_cd"`
	Msg1   string                 `json:"msg1"`
	Output ProductBasicInfoOutput `json:"output"`
}

// ProductBasicInfoOutput represents generic product basic info payload.
type ProductBasicInfoOutput struct {
	PdNo               string `json:"pdno"`
	PrdtTypeCD         string `json:"prdt_type_cd"`
	PrdtName           string `json:"prdt_name"`
	PrdtName120        string `json:"prdt_name120"`
	PrdtAbrvName       string `json:"prdt_abrv_name"`
	PrdtEngName        string `json:"prdt_eng_name"`
	PrdtEngName120     string `json:"prdt_eng_name120"`
	PrdtEngAbrvName    string `json:"prdt_eng_abrv_name"`
	StdPdNo            string `json:"std_pdno"`
	ShtnPdNo           string `json:"shtn_pdno"`
	PrdtSaleStatCD     string `json:"prdt_sale_stat_cd"`
	PrdtRiskGradCD     string `json:"prdt_risk_grad_cd"`
	PrdtClsfCD         string `json:"prdt_clsf_cd"`
	PrdtClsfName       string `json:"prdt_clsf_name"`
	SaleStrtDt         string `json:"sale_strt_dt"`
	SaleEndDt          string `json:"sale_end_dt"`
	WrapAsstTypeCD     string `json:"wrap_asst_type_cd"`
	IvstPrdtTypeCD     string `json:"ivst_prdt_type_cd"`
	IvstPrdtTypeCDName string `json:"ivst_prdt_type_cd_name"`
	FrstErlmDt         string `json:"frst_erlm_dt"`
}

// OverseasProductBasicInfoResponse represents overseas product basic info response.
type OverseasProductBasicInfoResponse struct {
	RtCD   string                         `json:"rt_cd"`
	MsgCD  string                         `json:"msg_cd"`
	Msg1   string                         `json:"msg1"`
	Output OverseasProductBasicInfoOutput `json:"output"`
}

// OverseasProductBasicInfoOutput represents overseas product basic info payload.
type OverseasProductBasicInfoOutput struct {
	StdPdNo                string `json:"std_pdno"`
	PrdtName               string `json:"prdt_name"`
	PrdtEngName            string `json:"prdt_eng_name"`
	OvrsItemName           string `json:"ovrs_item_name"`
	PrdtClsfCD             string `json:"prdt_clsf_cd"`
	PrdtClsfName           string `json:"prdt_clsf_name"`
	NatnCD                 string `json:"natn_cd"`
	NatnName               string `json:"natn_name"`
	TrMketCD               string `json:"tr_mket_cd"`
	TrMketName             string `json:"tr_mket_name"`
	OvrsExcgCD             string `json:"ovrs_excg_cd"`
	OvrsExcgName           string `json:"ovrs_excg_name"`
	TrCrcyCD               string `json:"tr_crcy_cd"`
	CrcyName               string `json:"crcy_name"`
	OvrsStckDvsnCD         string `json:"ovrs_stck_dvsn_cd"`
	LstgYN                 string `json:"lstg_yn"`
	LstgDt                 string `json:"lstg_dt"`
	LstgAbolItemYN         string `json:"lstg_abol_item_yn"`
	LstgAbolDt             string `json:"lstg_abol_dt"`
	OvrsStckTrStopDvsnCD   string `json:"ovrs_stck_tr_stop_dvsn_cd"`
	OvrsStckStopRsonCD     string `json:"ovrs_stck_stop_rson_cd"`
	DtmTrPsblYN            string `json:"dtm_tr_psbl_yn"`
	MemoText1              string `json:"memo_text1"`
	OvrsNowPric1           string `json:"ovrs_now_pric1"`
	LastRcvgDtime          string `json:"last_rcvg_dtime"`
	MiniStkTrStatDvsnCD    string `json:"mini_stk_tr_stat_dvsn_cd"`
	MintDcptTradPsblYN     string `json:"mint_dcpt_trad_psbl_yn"`
	MintFnumTradPsblYN     string `json:"mint_fnum_trad_psbl_yn"`
	PtpItemYN              string `json:"ptp_item_yn"`
	PtpItemTrfxExmtYN      string `json:"ptp_item_trfx_exmt_yn"`
	PtpItemTrfxExmtStrtDt  string `json:"ptp_item_trfx_exmt_strt_dt"`
	PtpItemTrfxExmtEndDt   string `json:"ptp_item_trfx_exmt_end_dt"`
	SdrfStopEclsYN         string `json:"sdrf_stop_ecls_yn"`
	SdrfStopEclsErlmDt     string `json:"sdrf_stop_ecls_erlm_dt"`
	PrdtTypeCD2            string `json:"prdt_type_cd_2"`
	OvrsStckPrdtGrpNo      string `json:"ovrs_stck_prdt_grp_no"`
	OvrsStckErlmRosnCD     string `json:"ovrs_stck_erlm_rosn_cd"`
	OvrsStckHistRghtDvsnCD string `json:"ovrs_stck_hist_rght_dvsn_cd"`
}

// ====================
// Order
// ====================

// OrderRequest represents KIS order request body
type OrderRequest struct {
	CANO         string `json:"CANO"`
	ACNT_PRDT_CD string `json:"ACNT_PRDT_CD"`
	PDNO         string `json:"PDNO"`
	ORD_DVSN     string `json:"ORD_DVSN"`
	ORD_QTY      string `json:"ORD_QTY"`
	ORD_UNPR     string `json:"ORD_UNPR"`
	EXCG_ID_DVSN string `json:"EXCG_ID_DVSN_CD,omitempty"`
	SLL_TYPE     string `json:"SLL_TYPE,omitempty"`
	CNDT_PRIC    string `json:"CNDT_PRIC,omitempty"`
}

// (OrderResponse is defined in orders.go)

// OrderRvseCnclRequest represents KIS order revise/cancel request body.
type OrderRvseCnclRequest struct {
	CANO           string `json:"CANO"`
	ACNT_PRDT_CD   string `json:"ACNT_PRDT_CD"`
	KRXFwdOrdOrgNo string `json:"KRX_FWDG_ORD_ORGNO"`
	OrgnODNo       string `json:"ORGN_ODNO"`
	OrdDvsn        string `json:"ORD_DVSN"`
	RvseCnclDvsnCD string `json:"RVSE_CNCL_DVSN_CD"`
	OrdQty         string `json:"ORD_QTY"`
	OrdUNPR        string `json:"ORD_UNPR"`
	QtyAllOrdYN    string `json:"QTY_ALL_ORD_YN"`
	ExcgIDDvsnCD   string `json:"EXCG_ID_DVSN_CD,omitempty"`
	CndtPric       string `json:"CNDT_PRIC,omitempty"`
}

// OverseasOrderRequest represents KIS overseas stock order request body.
type OverseasOrderRequest struct {
	CANO            string `json:"CANO"`
	ACNT_PRDT_CD    string `json:"ACNT_PRDT_CD"`
	OVRS_EXCG_CD    string `json:"OVRS_EXCG_CD"`
	PDNO            string `json:"PDNO"`
	ORD_QTY         string `json:"ORD_QTY"`
	OVRS_ORD_UNPR   string `json:"OVRS_ORD_UNPR"`
	CTAC_TLNO       string `json:"CTAC_TLNO,omitempty"`
	MGCO_APTM_ODNO  string `json:"MGCO_APTM_ODNO,omitempty"`
	SLL_TYPE        string `json:"SLL_TYPE,omitempty"`
	ORD_SVR_DVSN_CD string `json:"ORD_SVR_DVSN_CD,omitempty"`
	ORD_DVSN        string `json:"ORD_DVSN"`
}

// OverseasOrderRvseCnclRequest represents KIS overseas stock revise/cancel request body.
type OverseasOrderRvseCnclRequest struct {
	CANO              string `json:"CANO"`
	ACNT_PRDT_CD      string `json:"ACNT_PRDT_CD"`
	OVRS_EXCG_CD      string `json:"OVRS_EXCG_CD"`
	PDNO              string `json:"PDNO"`
	ORGN_ODNO         string `json:"ORGN_ODNO"`
	RVSE_CNCL_DVSN_CD string `json:"RVSE_CNCL_DVSN_CD"`
	ORD_QTY           string `json:"ORD_QTY"`
	OVRS_ORD_UNPR     string `json:"OVRS_ORD_UNPR"`
	MGCO_APTM_ODNO    string `json:"MGCO_APTM_ODNO,omitempty"`
	ORD_SVR_DVSN_CD   string `json:"ORD_SVR_DVSN_CD,omitempty"`
}

// StockRvseCnclCandidate represents one modifiable/cancellable order row.
type StockRvseCnclCandidate struct {
	ODNo         string `json:"odno"`
	OrgnODNo     string `json:"orgn_odno"`
	OrdGnoBrno   string `json:"ord_gno_brno"`
	OrdDvsn      string `json:"ord_dvsn"`
	OrdQty       string `json:"ord_qty"`
	OrdUNPR      string `json:"ord_unpr"`
	PsblQty      string `json:"psbl_qty"`
	ExcgIDDvsnCD string `json:"excg_id_dvsn_cd"`
}

// StockRvseCnclResponse represents KIS response for inquire-psbl-rvsecncl.
type StockRvseCnclResponse struct {
	RtCD         string                   `json:"rt_cd"`
	MsgCD        string                   `json:"msg_cd"`
	Msg1         string                   `json:"msg1"`
	Output       []StockRvseCnclCandidate `json:"output"`
	CtxAreaFK100 string                   `json:"ctx_area_fk100"`
	CtxAreaNK100 string                   `json:"ctx_area_nk100"`
}

// DomesticDailyCcldResponse represents domestic daily order/filled inquiry response.
type DomesticDailyCcldResponse struct {
	RtCD         string                  `json:"rt_cd"`
	MsgCD        string                  `json:"msg_cd"`
	Msg1         string                  `json:"msg1"`
	Output1      []DomesticDailyCcldItem `json:"output1"`
	CtxAreaFK100 string                  `json:"ctx_area_fk100"`
	CtxAreaNK100 string                  `json:"ctx_area_nk100"`
}

// DomesticDailyCcldItem represents one domestic order row.
type DomesticDailyCcldItem struct {
	OrdDt      string `json:"ord_dt"`
	OrdGnoBrno string `json:"ord_gno_brno"`
	ODNo       string `json:"odno"`
	OrgnODNo   string `json:"orgn_odno"`
	OrdTmd     string `json:"ord_tmd"`
	PdNo       string `json:"pdno"`
	PrdtName   string `json:"prdt_name"`
	OrdQty     string `json:"ord_qty"`
	OrdUNPR    string `json:"ord_unpr"`
	SllBuyDvsn string `json:"sll_buy_dvsn_cd"`
	TotCcldQty string `json:"tot_ccld_qty"`
	TotCcldAmt string `json:"tot_ccld_amt"`
	AvgPrvs    string `json:"avg_prvs"`
	CnclYN     string `json:"cncl_yn"`
	CncCfrmQty string `json:"cnc_cfrm_qty"`
	RmnQty     string `json:"rmn_qty"`
	RjctQty    string `json:"rjct_qty"`
	OrdDvsnCD  string `json:"ord_dvsn_cd"`
	ExcgIDDvsn string `json:"excg_id_dvsn_cd"`
}

// OverseasCcnlResponse represents overseas order/fill inquiry response.
type OverseasCcnlResponse struct {
	RtCD         string             `json:"rt_cd"`
	MsgCD        string             `json:"msg_cd"`
	Msg1         string             `json:"msg1"`
	CtxAreaFK200 string             `json:"ctx_area_fk200"`
	CtxAreaNK200 string             `json:"ctx_area_nk200"`
	Output       []OverseasCcnlItem `json:"output"`
}

// OverseasCcnlItem represents one overseas order row.
type OverseasCcnlItem struct {
	OrdDt         string `json:"ord_dt"`
	OrdTmd        string `json:"ord_tmd"`
	OrdGnoBrno    string `json:"ord_gno_brno"`
	ODNo          string `json:"odno"`
	OrgnODNo      string `json:"orgn_odno"`
	SllBuyDvsnCD  string `json:"sll_buy_dvsn_cd"`
	RvseCnclDvsn  string `json:"rvse_cncl_dvsn"`
	PdNo          string `json:"pdno"`
	PrdtName      string `json:"prdt_name"`
	FtOrdQty      string `json:"ft_ord_qty"`
	FtOrdUNPR3    string `json:"ft_ord_unpr3"`
	FtCcldQty     string `json:"ft_ccld_qty"`
	FtCcldUNPR3   string `json:"ft_ccld_unpr3"`
	FtCcldAmt3    string `json:"ft_ccld_amt3"`
	NccsQty       string `json:"nccs_qty"`
	PrcsStatName  string `json:"prcs_stat_name"`
	RjctRson      string `json:"rjct_rson"`
	RjctRsonName  string `json:"rjct_rson_name"`
	OvrsExcgCD    string `json:"ovrs_excg_cd"`
	TrCrcyCD      string `json:"tr_crcy_cd"`
	UsaAmkExtsYN  string `json:"usa_amk_exts_rqst_yn"`
	SpltBuyAttrNm string `json:"splt_buy_attr_name"`
}

// ====================
// Common Response
// ====================

// ErrorResponse represents KIS API error response
type ErrorResponse struct {
	RtCD  string `json:"rt_cd"`
	MsgCD string `json:"msg_cd"`
	Msg1  string `json:"msg1"`
}

// IsSuccess checks if the response indicates success
func (e *ErrorResponse) IsSuccess() bool {
	return e.RtCD == "0"
}

// ParseKISDate parses KIS date strings (YYYYMMDD)
func ParseKISDate(s string) (time.Time, error) {
	return time.Parse("20060102", s)
}

// ParseKISDateTime parses KIS datetime strings
func ParseKISDateTime(date, t string) (time.Time, error) {
	return time.Parse("20060102150405", date+t)
}
