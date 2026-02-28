package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/smallfish06/krsec/internal/kis"
	"github.com/smallfish06/krsec/pkg/broker"
)

type endpointDispatchFunc func(ctx context.Context, trID string, fields map[string]string) (map[string]interface{}, error)

type endpointDispatcher struct {
	adapter *Adapter
	routes  map[string]endpointRoute
}

type endpointRoute struct {
	methods map[string]struct{}
	fn      endpointDispatchFunc
}

func newEndpointRoute(methods []string, fn endpointDispatchFunc) endpointRoute {
	m := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		n := strings.ToUpper(strings.TrimSpace(method))
		if n == "" {
			continue
		}
		m[n] = struct{}{}
	}
	return endpointRoute{methods: m, fn: fn}
}

func (r endpointRoute) allows(method string) bool {
	_, ok := r.methods[strings.ToUpper(strings.TrimSpace(method))]
	return ok
}

func newEndpointDispatcher(adapter *Adapter) *endpointDispatcher {
	d := &endpointDispatcher{adapter: adapter}
	d.routes = map[string]endpointRoute{
		kis.PathDomesticStockInquirePrice:                    newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquirePrice),
		kis.PathDomesticStockInquireDailyPrice:               newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireDailyPrice),
		kis.PathDomesticStockInquireAskingPriceExpCcn:        newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireAskingPriceExpCcn),
		kis.PathDomesticStockInquireCcnl:                     newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireCcnl),
		kis.PathDomesticStockInquireTimeItemConclusion:       newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireTimeItemConclusion),
		kis.PathDomesticStockInquireMember:                   newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireMember),
		kis.PathETFETNComponentStockPrice:                    newEndpointRoute([]string{http.MethodGet}, d.dispatchETFETNComponentStockPrice),
		kis.PathDomesticStockVolumeRank:                      newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockVolumeRank),
		kis.PathDomesticStockRankingMarketCap:                newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockRankingFluctuation),
		kis.PathDomesticStockRankingFluctuation:              newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockRankingFluctuation),
		kis.PathDomesticStockInquireIndexPrice:               newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireIndexPrice),
		kis.PathDomesticStockInquireIndexDailyPrice:          newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireIndexDailyPrice),
		kis.PathDomesticStockInquireDailyIndexChart:          newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockInquireDailyIndexChart),
		kis.PathOverseasPricePrice:                           newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPricePrice),
		kis.PathOverseasPriceInquireDailyChartPrice:          newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPriceInquireDailyChartPrice),
		kis.PathOverseasPriceDailyPrice:                      newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPriceDailyPrice),
		kis.PathOverseasPricePriceDetail:                     newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPricePriceDetail),
		kis.PathOverseasPriceInquireCcnl:                     newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPriceInquireCcnl),
		kis.PathOverseasStockRankingUpdownRate:               newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasStockRankingUpdownRate),
		kis.PathOverseasPriceInquireTimeItemChart:            newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPriceInquireTimeItemChart),
		kis.PathDomesticBondInquirePrice:                     newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticBondInquirePrice),
		kis.PathDomesticBondInquireBalance:                   newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticBondInquireBalance),
		kis.PathDomesticStockSearchStockInfo:                 newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockSearchStockInfo),
		kis.PathDomesticStockSearchInfo:                      newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockSearchInfo),
		kis.PathOverseasPriceSearchInfo:                      newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasPriceSearchInfo),
		kis.PathDomesticStockTradingInquireBalance:           newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockTradingInquireBalance),
		kis.PathOverseasStockTradingInquireBalance:           newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasStockTradingInquireBalance),
		kis.PathOverseasStockTradingInquirePsAmount:          newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasStockTradingInquirePsAmount),
		kis.PathDomesticStockTradingInquirePsblOrder:         newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockTradingInquirePsblOrder),
		kis.PathDomesticStockTradingInquirePeriodTradeProfit: newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockTradingInquirePeriodTradeProfit),
		kis.PathDomesticStockTradingInquireDailyCcld:         newEndpointRoute([]string{http.MethodGet}, d.dispatchDomesticStockTradingInquireDailyCcld),
		kis.PathOverseasStockTradingInquireCcnl:              newEndpointRoute([]string{http.MethodGet}, d.dispatchOverseasStockTradingInquireCcnl),
		kis.PathDomesticStockTradingOrderCash:                newEndpointRoute([]string{http.MethodPost}, d.dispatchDomesticStockTradingOrderCash),
		kis.PathDomesticStockTradingOrderRvseCncl:            newEndpointRoute([]string{http.MethodPost}, d.dispatchDomesticStockTradingOrderRvseCncl),
		kis.PathOverseasStockTradingOrder:                    newEndpointRoute([]string{http.MethodPost}, d.dispatchOverseasStockTradingOrder),
		kis.PathOverseasStockTradingOrderRvseCncl:            newEndpointRoute([]string{http.MethodPost}, d.dispatchOverseasStockTradingOrderRvseCncl),
	}
	return d
}

func (d *endpointDispatcher) callEndpoint(
	ctx context.Context,
	method string,
	path string,
	trID string,
	fields map[string]string,
) (map[string]interface{}, error) {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		m = http.MethodGet
	}

	normalizedPath := normalizeEndpointPath(path)
	normalizedFields := normalizeEndpointFields(fields)

	route, ok := d.routes[normalizedPath]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported KIS endpoint path %s", broker.ErrInvalidOrderRequest, normalizedPath)
	}
	if !route.allows(m) {
		return nil, fmt.Errorf("%w: unsupported method %s", broker.ErrInvalidOrderRequest, m)
	}

	return route.fn(ctx, trID, normalizedFields)
}

// CallEndpoint dispatches a KIS endpoint path to implemented client methods.
func (a *Adapter) CallEndpoint(
	ctx context.Context,
	method string,
	path string,
	trID string,
	fields map[string]string,
) (map[string]interface{}, error) {
	dispatcher := a.dispatcher
	if dispatcher == nil {
		dispatcher = newEndpointDispatcher(a)
	}
	return dispatcher.callEndpoint(ctx, method, path, trID, fields)
}

func (d *endpointDispatcher) dispatchDomesticStockInquirePrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	symbol := getField(p, "FID_INPUT_ISCD", "PDNO")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	market := "KRX"
	if strings.EqualFold(getField(p, "FID_COND_MRKT_DIV_CODE"), "Q") {
		market = "KOSDAQ"
	}
	resp, err := d.adapter.client.InquirePrice(ctx, market, symbol)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireDailyPrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	symbol := getField(p, "FID_INPUT_ISCD", "PDNO")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	market := "KRX"
	if strings.EqualFold(getField(p, "FID_COND_MRKT_DIV_CODE"), "Q") {
		market = "KOSDAQ"
	}
	adjust := !strings.EqualFold(getField(p, "FID_ORG_ADJ_PRC"), "1")
	resp, err := d.adapter.client.InquireDailyPrice(
		ctx,
		market,
		symbol,
		getField(p, "FID_INPUT_DATE_1"),
		getField(p, "FID_INPUT_DATE_2"),
		adjust,
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireAskingPriceExpCcn(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireAskingPriceExpCcn(ctx, getField(p, "FID_COND_MRKT_DIV_CODE"), getField(p, "FID_INPUT_ISCD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireCcnl(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireCcnl(ctx, getField(p, "FID_COND_MRKT_DIV_CODE"), getField(p, "FID_INPUT_ISCD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireTimeItemConclusion(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireTimeItemConclusion(
		ctx,
		getField(p, "FID_COND_MRKT_DIV_CODE"),
		getField(p, "FID_INPUT_ISCD"),
		getField(p, "FID_INPUT_HOUR_1"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireMember(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireMember(ctx, getField(p, "FID_COND_MRKT_DIV_CODE"), getField(p, "FID_INPUT_ISCD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchETFETNComponentStockPrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireComponentStockPrice(
		ctx,
		getField(p, "FID_COND_MRKT_DIV_CODE"),
		getField(p, "FID_INPUT_ISCD"),
		getField(p, "FID_COND_SCR_DIV_CODE"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockVolumeRank(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireVolumeRank(ctx, kis.VolumeRankParams{
		MarketDiv:       getField(p, "FID_COND_MRKT_DIV_CODE"),
		ScreenDiv:       getField(p, "FID_COND_SCR_DIV_CODE"),
		InputISCD:       getField(p, "FID_INPUT_ISCD"),
		DivClsCode:      getField(p, "FID_DIV_CLS_CODE"),
		BlngClsCode:     getField(p, "FID_BLNG_CLS_CODE"),
		TrgtClsCode:     getField(p, "FID_TRGT_CLS_CODE"),
		TrgtExlsClsCode: getField(p, "FID_TRGT_EXLS_CLS_CODE"),
		InputPrice1:     getField(p, "FID_INPUT_PRICE_1"),
		InputPrice2:     getField(p, "FID_INPUT_PRICE_2"),
		VolCnt:          getField(p, "FID_VOL_CNT"),
		InputDate1:      getField(p, "FID_INPUT_DATE_1"),
	})
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockRankingFluctuation(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireMarketCapRank(ctx, kis.MarketCapRankParams{
		InputPrice2:     getField(p, "FID_INPUT_PRICE_2"),
		MarketDiv:       getField(p, "FID_COND_MRKT_DIV_CODE"),
		ScreenDiv:       getField(p, "FID_COND_SCR_DIV_CODE"),
		DivClsCode:      getField(p, "FID_DIV_CLS_CODE"),
		InputISCD:       getField(p, "FID_INPUT_ISCD"),
		TrgtClsCode:     getField(p, "FID_TRGT_CLS_CODE"),
		TrgtExlsClsCode: getField(p, "FID_TRGT_EXLS_CLS_CODE"),
		InputPrice1:     getField(p, "FID_INPUT_PRICE_1"),
		VolCnt:          getField(p, "FID_VOL_CNT"),
	})
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireIndexPrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireIndexPrice(ctx, getField(p, "FID_COND_MRKT_DIV_CODE"), getField(p, "FID_INPUT_ISCD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireIndexDailyPrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireIndexDailyPrice(
		ctx,
		getField(p, "FID_PERIOD_DIV_CODE"),
		getField(p, "FID_COND_MRKT_DIV_CODE"),
		getField(p, "FID_INPUT_ISCD"),
		getField(p, "FID_INPUT_DATE_1"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockInquireDailyIndexChart(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireDailyIndexChartPrice(
		ctx,
		getField(p, "FID_COND_MRKT_DIV_CODE"),
		getField(p, "FID_INPUT_ISCD"),
		getField(p, "FID_INPUT_DATE_1"),
		getField(p, "FID_INPUT_DATE_2"),
		getField(p, "FID_PERIOD_DIV_CODE"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPricePrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasPrice(ctx, getField(p, "EXCD"), getField(p, "SYMB"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPriceInquireDailyChartPrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasDailyChartPrice(
		ctx,
		getField(p, "FID_COND_MRKT_DIV_CODE"),
		getField(p, "FID_INPUT_ISCD"),
		getField(p, "FID_INPUT_DATE_1"),
		getField(p, "FID_INPUT_DATE_2"),
		getField(p, "FID_PERIOD_DIV_CODE"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPriceDailyPrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasDailyPrice(
		ctx,
		getField(p, "AUTH"),
		getField(p, "EXCD"),
		getField(p, "SYMB"),
		getField(p, "GUBN"),
		getField(p, "BYMD"),
		getField(p, "MODP"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPricePriceDetail(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasPriceDetail(
		ctx,
		getField(p, "AUTH"),
		getField(p, "EXCD"),
		getField(p, "SYMB"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPriceInquireCcnl(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasTick(
		ctx,
		getField(p, "EXCD"),
		getField(p, "TDAY"),
		getField(p, "SYMB"),
		getField(p, "AUTH"),
		getField(p, "KEYB"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasStockRankingUpdownRate(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasUpdownRate(
		ctx,
		getField(p, "EXCD"),
		getField(p, "NDAY"),
		getField(p, "GUBN"),
		getField(p, "VOL_RANG"),
		getField(p, "AUTH"),
		getField(p, "KEYB"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPriceInquireTimeItemChart(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasTimeItemChartPrice(
		ctx,
		getField(p, "AUTH"),
		getField(p, "EXCD"),
		getField(p, "SYMB"),
		getField(p, "NMIN"),
		getField(p, "PINC"),
		getField(p, "NEXT"),
		getField(p, "NREC"),
		getField(p, "FILL"),
		getField(p, "KEYB"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticBondInquirePrice(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	isin := getField(p, "FID_INPUT_ISCD")
	if isin == "" {
		return nil, broker.ErrInvalidSymbol
	}
	startDate := getField(p, "FID_INPUT_DATE_1")
	endDate := getField(p, "FID_INPUT_DATE_2")
	if startDate != "" || endDate != "" {
		resp, err := d.adapter.client.InquireBondDaily(ctx, isin, startDate, endDate)
		return marshalMap(resp, err)
	}
	resp, err := d.adapter.client.InquireBondPrice(ctx, isin)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticBondInquireBalance(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquireBondBalance(ctx, cano, prdt)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockSearchStockInfo(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireStockBasicInfo(ctx, getField(p, "PDNO"), getField(p, "PRDT_TYPE_CD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockSearchInfo(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireProductBasicInfo(ctx, getField(p, "PDNO"), getField(p, "PRDT_TYPE_CD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasPriceSearchInfo(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOverseasProductBasicInfo(ctx, getField(p, "PDNO"), getField(p, "PRDT_TYPE_CD"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockTradingInquireBalance(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquireAccountBalance(ctx, cano, prdt)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasStockTradingInquireBalance(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquireOverseasBalanceRaw(
		ctx,
		cano, prdt,
		getField(p, "OVRS_EXCG_CD"),
		getField(p, "TR_CRCY_CD"),
		getField(p, "CTX_AREA_FK200"),
		getField(p, "CTX_AREA_NK200"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasStockTradingInquirePsAmount(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquireOverseasPsAmount(
		ctx,
		cano, prdt,
		getField(p, "OVRS_EXCG_CD"),
		getField(p, "OVRS_ORD_UNPR"),
		getField(p, "ITEM_CD"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockTradingInquirePsblOrder(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquirePossibleOrder(
		ctx,
		cano, prdt,
		getField(p, "PDNO"),
		getField(p, "ORD_UNPR"),
		getField(p, "ORD_DVSN"),
		getField(p, "CMA_EVLU_AMT_ICLD_YN"),
		getField(p, "OVRS_ICLD_YN"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockTradingInquirePeriodTradeProfit(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquirePeriodTradeProfit(
		ctx,
		cano, prdt,
		getField(p, "SORT_DVSN"),
		getField(p, "INQR_STRT_DT"),
		getField(p, "INQR_END_DT"),
		getField(p, "CBLC_DVSN"),
		getField(p, "PDNO"),
		getField(p, "CTX_AREA_NK100"),
		getField(p, "CTX_AREA_FK100"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockTradingInquireDailyCcld(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquireDailyCcld(
		ctx,
		cano, prdt,
		getField(p, "INQR_STRT_DT"),
		getField(p, "INQR_END_DT"),
		getField(p, "ORD_GNO_BRNO"),
		getField(p, "ODNO"),
		getField(p, "EXCG_ID_DVSN_CD"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasStockTradingInquireCcnl(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	resp, err := d.adapter.client.InquireOverseasCcnl(
		ctx,
		cano, prdt,
		getField(p, "ORD_STRT_DT"),
		getField(p, "ORD_END_DT"),
		getField(p, "OVRS_EXCG_CD"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockTradingOrderCash(ctx context.Context, trID string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	side := inferDomesticSide(trID, getField(p, "SIDE", "SLL_BUY_DVSN_CD"))
	orderType := inferOrderType(getField(p, "ORD_DVSN"))
	qty, err := atoiField(getField(p, "ORD_QTY"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ORD_QTY", broker.ErrInvalidOrderRequest)
	}
	price, err := atoiField(getField(p, "ORD_UNPR"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ORD_UNPR", broker.ErrInvalidOrderRequest)
	}
	resp, err := d.adapter.client.OrderCash(
		ctx,
		cano, prdt,
		getField(p, "PDNO"),
		orderType,
		qty,
		price,
		side,
		getField(p, "EXCG_ID_DVSN", "EXCG_ID_DVSN_CD"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticStockTradingOrderRvseCncl(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	qty, err := atoiField(getField(p, "ORD_QTY"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ORD_QTY", broker.ErrInvalidOrderRequest)
	}
	price, err := atoiField(getField(p, "ORD_UNPR"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ORD_UNPR", broker.ErrInvalidOrderRequest)
	}
	resp, err := d.adapter.client.OrderRvseCncl(
		ctx,
		cano, prdt,
		getField(p, "KRX_FWDG_ORD_ORGNO"),
		getField(p, "ORGN_ODNO"),
		getField(p, "ORD_DVSN"),
		getField(p, "RVSE_CNCL_DVSN_CD"),
		qty,
		price,
		parseBoolYN(getField(p, "QTY_ALL_ORD_YN")),
		getField(p, "EXCG_ID_DVSN_CD"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasStockTradingOrder(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	qty, err := atoiField(getField(p, "ORD_QTY"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ORD_QTY", broker.ErrInvalidOrderRequest)
	}
	price, err := atofField(getField(p, "OVRS_ORD_UNPR"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid OVRS_ORD_UNPR", broker.ErrInvalidOrderRequest)
	}
	resp, err := d.adapter.client.OrderOverseas(
		ctx,
		cano, prdt,
		getField(p, "OVRS_EXCG_CD"),
		getField(p, "PDNO"),
		qty,
		price,
		inferOverseasSide(getField(p, "SIDE", "SLL_TYPE")),
		getField(p, "ORD_DVSN"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOverseasStockTradingOrderRvseCncl(ctx context.Context, _ string, p map[string]string) (map[string]interface{}, error) {
	cano, prdt := d.adapter.resolveEndpointAccount(p)
	qty, err := atoiField(getField(p, "ORD_QTY"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ORD_QTY", broker.ErrInvalidOrderRequest)
	}
	price, err := atofField(getField(p, "OVRS_ORD_UNPR"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid OVRS_ORD_UNPR", broker.ErrInvalidOrderRequest)
	}
	resp, err := d.adapter.client.OrderOverseasRvseCncl(
		ctx,
		cano, prdt,
		getField(p, "OVRS_EXCG_CD"),
		getField(p, "PDNO"),
		getField(p, "ORGN_ODNO"),
		getField(p, "RVSE_CNCL_DVSN_CD"),
		qty,
		price,
	)
	return marshalMap(resp, err)
}

func normalizeEndpointPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(path, kis.PathPrefixUAPISlash) {
		path = kis.PathPrefixUAPI + path
	}
	return path
}

func normalizeEndpointFields(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		key := strings.ToUpper(strings.TrimSpace(k))
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(v)
	}
	return out
}

func getField(fields map[string]string, keys ...string) string {
	for _, k := range keys {
		key := strings.ToUpper(strings.TrimSpace(k))
		if key == "" {
			continue
		}
		if v, ok := fields[key]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (a *Adapter) resolveEndpointAccount(fields map[string]string) (string, string) {
	accountID := getField(fields, "ACCOUNT_ID")
	if accountID != "" {
		return a.parseAccountID(accountID)
	}
	cano := getField(fields, "CANO")
	prdt := getField(fields, "ACNT_PRDT_CD")
	if cano == "" {
		cano = a.accountID
	}
	if prdt == "" {
		prdt = a.accountPrdtCD
	}
	return cano, prdt
}

func atoiField(v string) (int, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, nil
	}
	return strconv.Atoi(v)
}

func atofField(v string) (float64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, nil
	}
	return strconv.ParseFloat(v, 64)
}

func parseBoolYN(v string) bool {
	v = strings.TrimSpace(strings.ToUpper(v))
	return v == "Y" || v == "1" || v == "TRUE"
}

func inferOrderType(ordDvsn string) string {
	if strings.TrimSpace(ordDvsn) == "01" {
		return "market"
	}
	return "limit"
}

func inferDomesticSide(trID, side string) string {
	if strings.EqualFold(strings.TrimSpace(side), "sell") || strings.TrimSpace(side) == "02" {
		return "sell"
	}
	switch strings.TrimSpace(strings.ToUpper(trID)) {
	case "TTTC0801U", "VTTC0801U":
		return "sell"
	default:
		return "buy"
	}
}

func inferOverseasSide(side string) string {
	s := strings.TrimSpace(strings.ToLower(side))
	if s == "sell" || s == "00" || s == "2" {
		return "sell"
	}
	return "buy"
}

func marshalMap(v interface{}, err error) (map[string]interface{}, error) {
	if err != nil {
		return nil, err
	}
	if v == nil {
		return map[string]interface{}{}, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	out := make(map[string]interface{})
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode response map: %w", err)
	}
	return out, nil
}
