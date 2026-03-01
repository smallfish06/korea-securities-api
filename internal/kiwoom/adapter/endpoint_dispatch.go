package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/smallfish06/krsec/internal/kiwoom"
	"github.com/smallfish06/krsec/pkg/broker"
)

type endpointDispatchFunc func(ctx context.Context, apiID string, fields map[string]string) (map[string]interface{}, error)

type endpointRouteKey struct {
	path  string
	apiID string
}

type endpointDispatcher struct {
	adapter *Adapter
	routes  map[endpointRouteKey]endpointRoute
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
	d.routes = map[endpointRouteKey]endpointRoute{
		{path: kiwoom.PathStockInfo, apiID: kiwoom.APIIDDomesticQuote}:              newEndpointRoute([]string{http.MethodPost}, d.dispatchDomesticQuote),
		{path: kiwoom.PathStockInfo, apiID: kiwoom.APIIDDomesticExecutionInfo}:      newEndpointRoute([]string{http.MethodPost}, d.dispatchDomesticExecutionInfo),
		{path: kiwoom.PathStockInfo, apiID: kiwoom.APIIDInstrumentInfo}:             newEndpointRoute([]string{http.MethodPost}, d.dispatchInstrumentInfo),
		{path: kiwoom.PathStockInfo, apiID: kiwoom.APIIDInvestorByStock}:            newEndpointRoute([]string{http.MethodPost}, d.dispatchInvestorByStock),
		{path: kiwoom.PathMarketCond, apiID: kiwoom.APIIDDomesticOrderBook}:         newEndpointRoute([]string{http.MethodPost}, d.dispatchDomesticOrderBook),
		{path: kiwoom.PathRankingInfo, apiID: kiwoom.APIIDVolumeRank}:               newEndpointRoute([]string{http.MethodPost}, d.dispatchVolumeRank),
		{path: kiwoom.PathRankingInfo, apiID: kiwoom.APIIDChangeRateRank}:           newEndpointRoute([]string{http.MethodPost}, d.dispatchChangeRateRank),
		{path: kiwoom.PathSector, apiID: kiwoom.APIIDSectorCurrent}:                 newEndpointRoute([]string{http.MethodPost}, d.dispatchSectorCurrent),
		{path: kiwoom.PathSector, apiID: kiwoom.APIIDSectorByPrice}:                 newEndpointRoute([]string{http.MethodPost}, d.dispatchSectorByPrice),
		{path: kiwoom.PathELW, apiID: kiwoom.APIIDELWDetail}:                        newEndpointRoute([]string{http.MethodPost}, d.dispatchELWDetail),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountBalance}:               newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountBalance),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountPositions}:             newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountPositions),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDUnsettledOrders}:              newEndpointRoute([]string{http.MethodPost}, d.dispatchUnsettledOrders),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDOrderExecutions}:              newEndpointRoute([]string{http.MethodPost}, d.dispatchOrderExecutions),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountDepositDetail}:         newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountDepositDetail),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountOrderExecutionDetail}:  newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountOrderExecutionDetail),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountOrderExecutionStatus}:  newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountOrderExecutionStatus),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountOrderableWithdrawable}: newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountOrderableWithdrawable),
		{path: kiwoom.PathAccount, apiID: kiwoom.APIIDAccountMarginDetail}:          newEndpointRoute([]string{http.MethodPost}, d.dispatchAccountMarginDetail),
		{path: kiwoom.PathChart, apiID: kiwoom.APIIDTickChart}:                      newEndpointRoute([]string{http.MethodPost}, d.dispatchTickChart),
		{path: kiwoom.PathChart, apiID: kiwoom.APIIDInvestorByStockChart}:           newEndpointRoute([]string{http.MethodPost}, d.dispatchInvestorByStockChart),
		{path: kiwoom.PathChart, apiID: kiwoom.APIIDDailyChart}:                     newEndpointRoute([]string{http.MethodPost}, d.dispatchDailyChart),
		{path: kiwoom.PathChart, apiID: kiwoom.APIIDWeeklyChart}:                    newEndpointRoute([]string{http.MethodPost}, d.dispatchWeeklyChart),
		{path: kiwoom.PathChart, apiID: kiwoom.APIIDMonthlyChart}:                   newEndpointRoute([]string{http.MethodPost}, d.dispatchMonthlyChart),
		{path: kiwoom.PathOrder, apiID: kiwoom.APIIDPlaceBuyOrder}:                  newEndpointRoute([]string{http.MethodPost}, d.dispatchPlaceOrder),
		{path: kiwoom.PathOrder, apiID: kiwoom.APIIDPlaceSellOrder}:                 newEndpointRoute([]string{http.MethodPost}, d.dispatchPlaceOrder),
		{path: kiwoom.PathOrder, apiID: kiwoom.APIIDModifyOrder}:                    newEndpointRoute([]string{http.MethodPost}, d.dispatchModifyOrder),
		{path: kiwoom.PathOrder, apiID: kiwoom.APIIDCancelOrder}:                    newEndpointRoute([]string{http.MethodPost}, d.dispatchCancelOrder),
	}
	d.registerDocumentedCustomRoutes()
	return d
}

func (d *endpointDispatcher) registerDocumentedCustomRoutes() {
	for _, key := range documentedCustomRouteKeys {
		if _, exists := d.routes[key]; exists {
			continue
		}
		d.routes[key] = newEndpointRoute([]string{http.MethodPost}, d.dispatchDocumentedEndpoint(key.path))
	}
}

func (d *endpointDispatcher) dispatchDocumentedEndpoint(path string) endpointDispatchFunc {
	return func(ctx context.Context, apiID string, fields map[string]string) (map[string]interface{}, error) {
		if d.adapter == nil || d.adapter.client == nil {
			return nil, fmt.Errorf("%w: kiwoom client is not initialized", broker.ErrInvalidOrderRequest)
		}
		payload := fieldsToBody(fields)
		applyDocumentedDefaults(apiID, payload)
		resp, err := d.adapter.client.CallDocumentedEndpoint(ctx, apiID, path, payload)
		return marshalMap(resp, err)
	}
}

func applyDocumentedDefaults(apiID string, payload map[string]interface{}) {
	switch strings.ToLower(strings.TrimSpace(apiID)) {
	case "ka50079", "ka50080":
		if payloadFieldEmpty(payload, "tic_scope") {
			payload["tic_scope"] = "1"
		}
	}
}

func payloadFieldEmpty(payload map[string]interface{}, key string) bool {
	if payload == nil {
		return true
	}
	v, ok := payload[key]
	if !ok || v == nil {
		return true
	}
	return strings.TrimSpace(fmt.Sprint(v)) == ""
}

// CallEndpoint dispatches a Kiwoom endpoint path/api_id to implemented client methods.
func (a *Adapter) CallEndpoint(
	ctx context.Context,
	method string,
	path string,
	apiID string,
	fields map[string]string,
) (map[string]interface{}, error) {
	dispatcher := a.dispatcher
	if dispatcher == nil {
		dispatcher = newEndpointDispatcher(a)
	}
	return dispatcher.callEndpoint(ctx, method, path, apiID, fields)
}

func (d *endpointDispatcher) callEndpoint(
	ctx context.Context,
	method string,
	path string,
	apiID string,
	fields map[string]string,
) (map[string]interface{}, error) {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		m = http.MethodPost
	}

	normalizedPath := normalizeEndpointPath(path)
	normalizedAPIID := normalizeEndpointAPIID(apiID)
	normalizedFields := normalizeEndpointFields(fields)

	if normalizedAPIID == "" {
		return nil, fmt.Errorf("%w: api_id is required", broker.ErrInvalidOrderRequest)
	}

	route, ok := d.routes[endpointRouteKey{path: normalizedPath, apiID: normalizedAPIID}]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported Kiwoom endpoint path/api_id %s/%s", broker.ErrInvalidOrderRequest, normalizedPath, normalizedAPIID)
	}
	if !route.allows(m) {
		return nil, fmt.Errorf("%w: unsupported method %s", broker.ErrInvalidOrderRequest, m)
	}

	return route.fn(ctx, normalizedAPIID, normalizedFields)
}

func (d *endpointDispatcher) dispatchDomesticQuote(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquirePrice(ctx, symbol)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticExecutionInfo(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireExecutionInfo(ctx, symbol)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDomesticOrderBook(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireOrderBook(ctx, symbol)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchInstrumentInfo(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireInstrumentInfo(ctx, symbol)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchInvestorByStock(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireInvestorByStock(ctx, symbol, fieldsToBody(fields, "STK_CD", "SYMBOL"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchSectorCurrent(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireSectorCurrent(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchSectorByPrice(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireSectorByPrice(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchVolumeRank(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireVolumeRank(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchChangeRateRank(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireChangeRateRank(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchELWDetail(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireELWDetail(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountBalance(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireBalance(ctx, getField(fields, "DMST_STEX_TP", "EXCHANGE"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountPositions(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquirePositions(
		ctx,
		getField(fields, "QRY_TP"),
		getField(fields, "DMST_STEX_TP", "EXCHANGE"),
	)
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchUnsettledOrders(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireUnsettledOrders(ctx, getField(fields, "STK_CD", "SYMBOL"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchOrderExecutions(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOrderExecutions(ctx, getField(fields, "STK_CD", "SYMBOL"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountDepositDetail(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireDepositDetail(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountOrderExecutionDetail(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOrderExecutionDetail(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountOrderExecutionStatus(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOrderExecutionStatus(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountOrderableWithdrawable(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireOrderableWithdrawable(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchAccountMarginDetail(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	resp, err := d.adapter.client.InquireMarginDetail(ctx, fieldsToBody(fields))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchTickChart(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireTickChart(ctx, symbol, getField(fields, "BASE_DT", "BASE_DATE"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchInvestorByStockChart(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireInvestorByStockChart(ctx, symbol, fieldsToBody(fields, "STK_CD", "SYMBOL"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchDailyChart(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireDailyPrice(ctx, symbol, getField(fields, "BASE_DT", "BASE_DATE"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchWeeklyChart(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireWeeklyPrice(ctx, symbol, getField(fields, "BASE_DT", "BASE_DATE"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchMonthlyChart(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	resp, err := d.adapter.client.InquireMonthlyPrice(ctx, symbol, getField(fields, "BASE_DT", "BASE_DATE"))
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchPlaceOrder(ctx context.Context, apiID string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	qty, err := parseInt64Field(getField(fields, "ORD_QTY", "QUANTITY"))
	if err != nil || qty <= 0 {
		return nil, fmt.Errorf("%w: invalid ORD_QTY", broker.ErrInvalidOrderRequest)
	}

	side := kiwoom.StockOrderSideBuy
	if apiID == kiwoom.APIIDPlaceSellOrder {
		side = kiwoom.StockOrderSideSell
	}
	resp, err := d.adapter.client.PlaceStockOrder(ctx, kiwoom.PlaceStockOrderRequest{
		Side:           side,
		Exchange:       getField(fields, "DMST_STEX_TP", "EXCHANGE"),
		Symbol:         symbol,
		Quantity:       qty,
		OrderPrice:     getField(fields, "ORD_UV", "ORDER_PRICE"),
		TradeType:      getField(fields, "TRDE_TP", "TRADE_TYPE"),
		ConditionPrice: getField(fields, "COND_UV", "CONDITION_PRICE"),
	})
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchModifyOrder(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	qty, err := parseInt64Field(getField(fields, "MDFY_QTY", "MODIFY_QTY"))
	if err != nil || qty <= 0 {
		return nil, fmt.Errorf("%w: invalid MDFY_QTY", broker.ErrInvalidOrderRequest)
	}
	resp, err := d.adapter.client.ModifyStockOrder(ctx, kiwoom.ModifyStockOrderRequest{
		Exchange:       getField(fields, "DMST_STEX_TP", "EXCHANGE"),
		OriginalID:     getField(fields, "ORIG_ORD_NO", "ORIGINAL_ORDER_ID"),
		Symbol:         symbol,
		ModifyQty:      qty,
		ModifyPrice:    getField(fields, "MDFY_UV", "MODIFY_PRICE"),
		ConditionPrice: getField(fields, "MDFY_COND_UV", "MODIFY_CONDITION_PRICE"),
	})
	return marshalMap(resp, err)
}

func (d *endpointDispatcher) dispatchCancelOrder(ctx context.Context, _ string, fields map[string]string) (map[string]interface{}, error) {
	symbol := getField(fields, "STK_CD", "SYMBOL")
	if symbol == "" {
		return nil, broker.ErrInvalidSymbol
	}
	qty, err := parseInt64Field(getField(fields, "CNCL_QTY", "CANCEL_QTY"))
	if err != nil || qty <= 0 {
		return nil, fmt.Errorf("%w: invalid CNCL_QTY", broker.ErrInvalidOrderRequest)
	}
	resp, err := d.adapter.client.CancelStockOrder(ctx, kiwoom.CancelStockOrderRequest{
		Exchange:   getField(fields, "DMST_STEX_TP", "EXCHANGE"),
		OriginalID: getField(fields, "ORIG_ORD_NO", "ORIGINAL_ORDER_ID"),
		Symbol:     symbol,
		CancelQty:  qty,
	})
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
	if !strings.HasPrefix(path, kiwoom.PathPrefixAPISlash) {
		path = kiwoom.PathPrefixAPI + path
	}
	return path
}

func normalizeEndpointAPIID(apiID string) string {
	return strings.ToLower(strings.TrimSpace(apiID))
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

func fieldsToBody(fields map[string]string, excludeKeys ...string) map[string]interface{} {
	excluded := make(map[string]struct{}, len(excludeKeys))
	for _, key := range excludeKeys {
		n := strings.ToUpper(strings.TrimSpace(key))
		if n != "" {
			excluded[n] = struct{}{}
		}
	}
	out := make(map[string]interface{}, len(fields))
	for k, v := range fields {
		if _, skip := excluded[k]; skip {
			continue
		}
		out[strings.ToLower(k)] = v
	}
	return out
}

func parseInt64Field(v string) (int64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, nil
	}
	return strconv.ParseInt(v, 10, 64)
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

	// Most Kiwoom responses are JSON objects. Some typed wrappers return arrays;
	// wrap those as {"items": [...]} to keep CallEndpoint's map contract.
	out := make(map[string]interface{})
	if err := json.Unmarshal(data, &out); err == nil {
		return out, nil
	}

	items := make([]interface{}, 0)
	if err := json.Unmarshal(data, &items); err == nil {
		return map[string]interface{}{"items": items}, nil
	}

	return nil, fmt.Errorf("decode response map: expected object or array")
}
