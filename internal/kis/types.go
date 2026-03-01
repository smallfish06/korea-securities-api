package kis

import (
	"time"

	kisspecs "github.com/smallfish06/krsec/internal/kis/specs"
)

// Legacy KIS DTO aliases. Keep these names stable while sourcing fields from
// generated documented endpoint types.
type StockPriceOutput = kisspecs.KISDomesticStockV1QuotationsInquirePriceOutputItem

type StockDailyPriceOutput = kisspecs.KISDomesticStockV1QuotationsInquireDailyPriceOutputItem

type StockBalanceOutput = kisspecs.KISDomesticStockV1TradingInquireBalanceOutput1Item

type StockBalanceSummary = kisspecs.KISDomesticStockV1TradingInquireBalanceOutput2Item

type BondPriceOutput = kisspecs.KISDomesticBondV1QuotationsInquirePriceOutputItem

type BondBalanceOutput = kisspecs.KISDomesticBondV1TradingInquireBalanceOutputItem

type OverseasPriceOutput = kisspecs.KISOverseasPriceV1QuotationsPriceDetailOutputItem

type StockBasicInfoOutput = kisspecs.KISDomesticStockV1QuotationsSearchStockInfoOutputItem

type ProductBasicInfoOutput = kisspecs.KISDomesticStockV1QuotationsSearchInfoOutputItem

type OverseasProductBasicInfoOutput = kisspecs.KISOverseasPriceV1QuotationsSearchInfoOutputItem

type StockRvseCnclCandidate = kisspecs.KISDomesticStockV1TradingInquirePsblRvsecnclOutputItem

type DomesticDailyCcldItem = kisspecs.KISDomesticStockV1TradingInquireDailyCcldOutput1Item

type OverseasCcnlItem = kisspecs.KISOverseasStockV1TradingInquireCcnlOutputItem

// ErrorResponse represents KIS API error response.
// Kept as a thin alias for compatibility.
type ErrorResponse = kisspecs.DocumentedResponseBase

// ParseKISDate parses KIS date strings (YYYYMMDD)
func ParseKISDate(s string) (time.Time, error) {
	return time.Parse("20060102", s)
}

// ParseKISDateTime parses KIS datetime strings.
func ParseKISDateTime(date, t string) (time.Time, error) {
	return time.Parse("20060102150405", date+t)
}
