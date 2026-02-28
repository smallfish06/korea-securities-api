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
		APIID:       APIIDDomesticQuote,
		Method:      http.MethodPost,
		Path:        PathStockInfo,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointDomesticExecutionInfo = endpointSpec{
		APIID:       APIIDDomesticExecutionInfo,
		Method:      http.MethodPost,
		Path:        PathStockInfo,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointDomesticOrderBook = endpointSpec{
		APIID:       APIIDDomesticOrderBook,
		Method:      http.MethodPost,
		Path:        PathMarketCond,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInstrumentInfo = endpointSpec{
		APIID:       APIIDInstrumentInfo,
		Method:      http.MethodPost,
		Path:        PathStockInfo,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInvestorByStock = endpointSpec{
		APIID:       APIIDInvestorByStock,
		Method:      http.MethodPost,
		Path:        PathStockInfo,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointSectorCurrent = endpointSpec{
		APIID:       APIIDSectorCurrent,
		Method:      http.MethodPost,
		Path:        PathSector,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointSectorByPrice = endpointSpec{
		APIID:       APIIDSectorByPrice,
		Method:      http.MethodPost,
		Path:        PathSector,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointVolumeRank = endpointSpec{
		APIID:       APIIDVolumeRank,
		Method:      http.MethodPost,
		Path:        PathRankingInfo,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointChangeRateRank = endpointSpec{
		APIID:       APIIDChangeRateRank,
		Method:      http.MethodPost,
		Path:        PathRankingInfo,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointELWDetail = endpointSpec{
		APIID:       APIIDELWDetail,
		Method:      http.MethodPost,
		Path:        PathELW,
		ContentType: "application/json;charset=UTF-8",
	}

	endpointAccountBalance = endpointSpec{
		APIID:       APIIDAccountBalance,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountPositions = endpointSpec{
		APIID:       APIIDAccountPositions,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointUnsettledOrders = endpointSpec{
		APIID:       APIIDUnsettledOrders,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointOrderExecutions = endpointSpec{
		APIID:       APIIDOrderExecutions,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountDepositDetail = endpointSpec{
		APIID:       APIIDAccountDepositDetail,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountOrderExecutionDetail = endpointSpec{
		APIID:       APIIDAccountOrderExecutionDetail,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountOrderExecutionStatus = endpointSpec{
		APIID:       APIIDAccountOrderExecutionStatus,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountOrderableWithdrawable = endpointSpec{
		APIID:       APIIDAccountOrderableWithdrawable,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointAccountMarginDetail = endpointSpec{
		APIID:       APIIDAccountMarginDetail,
		Method:      http.MethodPost,
		Path:        PathAccount,
		ContentType: "application/json;charset=UTF-8",
	}

	endpointTickChart = endpointSpec{
		APIID:       APIIDTickChart,
		Method:      http.MethodPost,
		Path:        PathChart,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointInvestorByStockChart = endpointSpec{
		APIID:       APIIDInvestorByStockChart,
		Method:      http.MethodPost,
		Path:        PathChart,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointDailyChart = endpointSpec{
		APIID:       APIIDDailyChart,
		Method:      http.MethodPost,
		Path:        PathChart,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointWeeklyChart = endpointSpec{
		APIID:       APIIDWeeklyChart,
		Method:      http.MethodPost,
		Path:        PathChart,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointMonthlyChart = endpointSpec{
		APIID:       APIIDMonthlyChart,
		Method:      http.MethodPost,
		Path:        PathChart,
		ContentType: "application/json;charset=UTF-8",
	}

	endpointPlaceBuyOrder = endpointSpec{
		APIID:       APIIDPlaceBuyOrder,
		Method:      http.MethodPost,
		Path:        PathOrder,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointPlaceSellOrder = endpointSpec{
		APIID:       APIIDPlaceSellOrder,
		Method:      http.MethodPost,
		Path:        PathOrder,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointModifyOrder = endpointSpec{
		APIID:       APIIDModifyOrder,
		Method:      http.MethodPost,
		Path:        PathOrder,
		ContentType: "application/json;charset=UTF-8",
	}
	endpointCancelOrder = endpointSpec{
		APIID:       APIIDCancelOrder,
		Method:      http.MethodPost,
		Path:        PathOrder,
		ContentType: "application/json;charset=UTF-8",
	}
)
