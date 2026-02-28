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
	endpointDomesticExecutionInfo = endpointSpec{
		APIID:       "ka10003",
		Method:      http.MethodPost,
		Path:        "/api/dostk/stkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointDomesticOrderBook = endpointSpec{
		APIID:       "ka10004",
		Method:      http.MethodPost,
		Path:        "/api/dostk/mrkcond",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInstrumentInfo = endpointSpec{
		APIID:       "ka10100",
		Method:      http.MethodPost,
		Path:        "/api/dostk/stkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInvestorByStock = endpointSpec{
		APIID:       "ka10059",
		Method:      http.MethodPost,
		Path:        "/api/dostk/stkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointSectorCurrent = endpointSpec{
		APIID:       "ka20001",
		Method:      http.MethodPost,
		Path:        "/api/dostk/sect",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointSectorByPrice = endpointSpec{
		APIID:       "ka20002",
		Method:      http.MethodPost,
		Path:        "/api/dostk/sect",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointVolumeRank = endpointSpec{
		APIID:       "ka10030",
		Method:      http.MethodPost,
		Path:        "/api/dostk/rkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointChangeRateRank = endpointSpec{
		APIID:       "ka10027",
		Method:      http.MethodPost,
		Path:        "/api/dostk/rkinfo",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointELWDetail = endpointSpec{
		APIID:       "ka30012",
		Method:      http.MethodPost,
		Path:        "/api/dostk/elw",
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
	endpointAccountDepositDetail = endpointSpec{
		APIID:       "kt00001",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountOrderExecutionDetail = endpointSpec{
		APIID:       "kt00007",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountOrderExecutionStatus = endpointSpec{
		APIID:       "kt00009",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountOrderableWithdrawable = endpointSpec{
		APIID:       "kt00010",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountMarginDetail = endpointSpec{
		APIID:       "kt00013",
		Method:      http.MethodPost,
		Path:        "/api/dostk/acnt",
		ContentType: "application/json;charset=UTF-8",
	}

	endpointTickChart = endpointSpec{
		APIID:       "ka10079",
		Method:      http.MethodPost,
		Path:        "/api/dostk/chart",
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInvestorByStockChart = endpointSpec{
		APIID:       "ka10060",
		Method:      http.MethodPost,
		Path:        "/api/dostk/chart",
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
