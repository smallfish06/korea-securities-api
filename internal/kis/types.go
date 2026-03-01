package kis

import "time"

// Legacy KIS DTO aliases. Keep these names stable while sourcing fields from
// generated documented endpoint types.
type StockPriceOutput = KISDomesticStockV1QuotationsInquirePriceOutputItem

type StockDailyPriceOutput = KISDomesticStockV1QuotationsInquireDailyPriceOutputItem

type StockBalanceOutput = KISDomesticStockV1TradingInquireBalanceOutput1Item

type StockBalanceSummary = KISDomesticStockV1TradingInquireBalanceOutput2Item

type BondPriceOutput = KISDomesticBondV1QuotationsInquirePriceOutputItem

type BondBalanceOutput = KISDomesticBondV1TradingInquireBalanceOutputItem

type OverseasPriceOutput = KISOverseasPriceV1QuotationsPriceDetailOutputItem

type StockBasicInfoOutput = KISDomesticStockV1QuotationsSearchStockInfoOutputItem

type ProductBasicInfoOutput = KISDomesticStockV1QuotationsSearchInfoOutputItem

type OverseasProductBasicInfoOutput = KISOverseasPriceV1QuotationsSearchInfoOutputItem

type StockRvseCnclCandidate = KISDomesticStockV1TradingInquirePsblRvsecnclOutputItem

type DomesticDailyCcldItem = KISDomesticStockV1TradingInquireDailyCcldOutput1Item

type OverseasCcnlItem = KISOverseasStockV1TradingInquireCcnlOutputItem

// ErrorResponse represents KIS API error response.
// Kept as a thin alias for compatibility.
type ErrorResponse = DocumentedResponseBase

// ParseKISDate parses KIS date strings (YYYYMMDD)
func ParseKISDate(s string) (time.Time, error) {
	return time.Parse("20060102", s)
}

// ParseKISDateTime parses KIS datetime strings.
func ParseKISDateTime(date, t string) (time.Time, error) {
	return time.Parse("20060102150405", date+t)
}
