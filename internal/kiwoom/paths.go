package kiwoom

// Kiwoom REST path prefixes.
const (
	PathPrefixAPI      = "/api"
	PathPrefixAPISlash = "/api/"
)

// Kiwoom endpoint paths.
const (
	PathStockInfo   = "/api/dostk/stkinfo"
	PathMarketCond  = "/api/dostk/mrkcond"
	PathForeignInst = "/api/dostk/frgnistt"
	PathRankingInfo = "/api/dostk/rkinfo"
	PathSector      = "/api/dostk/sect"
	PathELW         = "/api/dostk/elw"
	PathETF         = "/api/dostk/etf"
	PathTheme       = "/api/dostk/thme"
	PathSLB         = "/api/dostk/slb"
	PathShortSell   = "/api/dostk/shsa"
	PathAccount     = "/api/dostk/acnt"
	PathChart       = "/api/dostk/chart"
	PathOrder       = "/api/dostk/ordr"
	PathCreditOrder = "/api/dostk/crdordr"
	PathWebSocket   = "/api/dostk/websocket"
)

// Kiwoom API IDs used by implemented functions.
const (
	APIIDDomesticQuote                = "ka10001"
	APIIDDomesticExecutionInfo        = "ka10003"
	APIIDDomesticOrderBook            = "ka10004"
	APIIDInstrumentInfo               = "ka10100"
	APIIDInvestorByStock              = "ka10059"
	APIIDSectorCurrent                = "ka20001"
	APIIDSectorByPrice                = "ka20002"
	APIIDVolumeRank                   = "ka10030"
	APIIDChangeRateRank               = "ka10027"
	APIIDELWDetail                    = "ka30012"
	APIIDAccountBalance               = "kt00005"
	APIIDAccountPositions             = "kt00018"
	APIIDUnsettledOrders              = "ka10075"
	APIIDOrderExecutions              = "ka10076"
	APIIDAccountDepositDetail         = "kt00001"
	APIIDAccountOrderExecutionDetail  = "kt00007"
	APIIDAccountOrderExecutionStatus  = "kt00009"
	APIIDAccountOrderableWithdrawable = "kt00010"
	APIIDAccountMarginDetail          = "kt00013"
	APIIDTickChart                    = "ka10079"
	APIIDInvestorByStockChart         = "ka10060"
	APIIDDailyChart                   = "ka10081"
	APIIDWeeklyChart                  = "ka10082"
	APIIDMonthlyChart                 = "ka10083"
	APIIDPlaceBuyOrder                = "kt10000"
	APIIDPlaceSellOrder               = "kt10001"
	APIIDModifyOrder                  = "kt10002"
	APIIDCancelOrder                  = "kt10003"
)
