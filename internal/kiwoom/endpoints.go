package kiwoom

import "net/http"

// endpointSpec defines one concrete Kiwoom REST endpoint used by this client.
// We keep this static and typed (KIS-style) instead of looking up runtime catalog metadata.
type endpointSpec struct {
	APIID       string
	Method      string
	Path        string
	ContentType string
}

var (
	endpointDomesticQuote = endpointSpec{
		APIID:       "ka10001",
		Method:      http.MethodPost,
		Path:        "/api/dostk/stkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInstrumentInfo = endpointSpec{
		APIID:       "ka10100",
		Method:      http.MethodPost,
		Path:        "/api/dostk/stkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountBalance = endpointSpec{
		APIID:       "kt00005",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountPositions = endpointSpec{
		APIID:       "kt00018",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointUnsettledOrders = endpointSpec{
		APIID:       "ka10075",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointOrderExecutions = endpointSpec{
		APIID:       "ka10076",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointDailyChart = endpointSpec{
		APIID:       "ka10081",
		Method:      http.MethodPost,
		Path:        "/api/dostk/chart",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointWeeklyChart = endpointSpec{
		APIID:       "ka10082",
		Method:      http.MethodPost,
		Path:        "/api/dostk/chart",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointMonthlyChart = endpointSpec{
		APIID:       "ka10083",
		Method:      http.MethodPost,
		Path:        "/api/dostk/chart",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointPlaceBuyOrder = endpointSpec{
		APIID:       "kt10000",
		Method:      http.MethodPost,
		Path:        "/api/dostk/ordr",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointPlaceSellOrder = endpointSpec{
		APIID:       "kt10001",
		Method:      http.MethodPost,
		Path:        "/api/dostk/ordr",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointModifyOrder = endpointSpec{
		APIID:       "kt10002",
		Method:      http.MethodPost,
		Path:        "/api/dostk/ordr",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointCancelOrder = endpointSpec{
		APIID:       "kt10003",
		Method:      http.MethodPost,
		Path:        "/api/dostk/ordr",
		ContentType: "application/json;charset=UTF-8",
	}
)
