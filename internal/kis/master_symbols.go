package kis

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

// MasterSymbol is a parsed symbol record from KIS master files (.mst/.cod).
type MasterSymbol struct {
	Symbol          string
	Market          string
	Name            string
	NameEn          string
	Exchange        string
	Currency        string
	Country         string
	ProductType     string
	ProductTypeCode string
	SecurityGroup   string
	IsListed        bool
}

type masterSymbolsIndex struct {
	byMarketSymbol map[string]MasterSymbol
	domesticBySym  map[string]MasterSymbol
	usBySym        map[string]MasterSymbol
	anyBySym       map[string]MasterSymbol
}

var (
	masterSymbolsMu            sync.RWMutex
	masterSymbolsBootstrapping bool
	masterSymbolsLoaded        bool
	masterSymbolsCount         int
	masterSymbols              = masterSymbolsIndex{
		byMarketSymbol: make(map[string]MasterSymbol),
		domesticBySym:  make(map[string]MasterSymbol),
		usBySym:        make(map[string]MasterSymbol),
		anyBySym:       make(map[string]MasterSymbol),
	}
)

type domesticMasterSource struct {
	URL      string
	Market   string
	Exchange string
	TailLen  int
}

type overseasMasterSource struct {
	URL      string
	Market   string
	Exchange string
	Country  string
}

var domesticMasterSources = []domesticMasterSource{
	{URL: "https://new.real.download.dws.co.kr/common/master/kospi_code.mst.zip", Market: "KOSPI", Exchange: "KRX", TailLen: 228},
	{URL: "https://new.real.download.dws.co.kr/common/master/kosdaq_code.mst.zip", Market: "KOSDAQ", Exchange: "KRX", TailLen: 222},
	{URL: "https://new.real.download.dws.co.kr/common/master/konex_code.mst.zip", Market: "KONEX", Exchange: "KRX", TailLen: 184},
	{URL: "https://new.real.download.dws.co.kr/common/master/nxt_kospi_code.mst.zip", Market: "NXT", Exchange: "NXT", TailLen: 228},
	{URL: "https://new.real.download.dws.co.kr/common/master/nxt_kosdaq_code.mst.zip", Market: "NXT", Exchange: "NXT", TailLen: 222},
}

var overseasMasterSources = []overseasMasterSource{
	{URL: "https://new.real.download.dws.co.kr/common/master/nasmst.cod.zip", Market: "US-NASDAQ", Exchange: "NASD", Country: "US"},
	{URL: "https://new.real.download.dws.co.kr/common/master/nysmst.cod.zip", Market: "US-NYSE", Exchange: "NYSE", Country: "US"},
	{URL: "https://new.real.download.dws.co.kr/common/master/amsmst.cod.zip", Market: "US-AMEX", Exchange: "AMEX", Country: "US"},
	{URL: "https://new.real.download.dws.co.kr/common/master/hksmst.cod.zip", Market: "HK", Exchange: "SEHK", Country: "HK"},
	{URL: "https://new.real.download.dws.co.kr/common/master/tsemst.cod.zip", Market: "JP", Exchange: "TKSE", Country: "JP"},
	{URL: "https://new.real.download.dws.co.kr/common/master/shsmst.cod.zip", Market: "SH", Exchange: "SHAA", Country: "CN"},
	{URL: "https://new.real.download.dws.co.kr/common/master/szsmst.cod.zip", Market: "SZ", Exchange: "SZAA", Country: "CN"},
	{URL: "https://new.real.download.dws.co.kr/common/master/hnxmst.cod.zip", Market: "HNX", Exchange: "HASE", Country: "VN"},
	{URL: "https://new.real.download.dws.co.kr/common/master/hsxmst.cod.zip", Market: "HSX", Exchange: "VNSE", Country: "VN"},
}

// BootstrapMasterSymbols downloads and parses KIS master symbol files once per process.
func (c *Client) BootstrapMasterSymbols(ctx context.Context) (int, error) {
	return c.rebuildMasterSymbols(ctx, false)
}

// ReloadMasterSymbols force-reloads master symbols even when already loaded.
func (c *Client) ReloadMasterSymbols(ctx context.Context) (int, error) {
	return c.rebuildMasterSymbols(ctx, true)
}

// LookupMasterSymbol returns a symbol record from preloaded master files.
func LookupMasterSymbol(market, symbol string) (MasterSymbol, bool) {
	masterSymbolsMu.RLock()
	idx := masterSymbols
	masterSymbolsMu.RUnlock()
	return lookupMasterSymbolFromIndex(idx, market, symbol)
}

func (c *Client) rebuildMasterSymbols(ctx context.Context, force bool) (int, error) {
	for {
		masterSymbolsMu.Lock()
		if !force && masterSymbolsLoaded {
			count := masterSymbolsCount
			masterSymbolsMu.Unlock()
			return count, nil
		}
		if !masterSymbolsBootstrapping {
			masterSymbolsBootstrapping = true
			masterSymbolsMu.Unlock()
			break
		}
		masterSymbolsMu.Unlock()

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	idx, count, err := buildMasterSymbolsIndex(ctx, c.httpClient)

	masterSymbolsMu.Lock()
	masterSymbolsBootstrapping = false
	if err != nil {
		masterSymbolsMu.Unlock()
		return 0, err
	}
	masterSymbols = idx
	masterSymbolsCount = count
	masterSymbolsLoaded = true
	masterSymbolsMu.Unlock()
	return count, nil
}

func buildMasterSymbolsIndex(ctx context.Context, httpClient *http.Client) (masterSymbolsIndex, int, error) {
	idx := masterSymbolsIndex{
		byMarketSymbol: make(map[string]MasterSymbol),
		domesticBySym:  make(map[string]MasterSymbol),
		usBySym:        make(map[string]MasterSymbol),
		anyBySym:       make(map[string]MasterSymbol),
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	total := 0

	for _, src := range domesticMasterSources {
		raw, err := downloadMasterPayload(ctx, httpClient, src.URL)
		if err != nil {
			continue
		}
		recs := parseDomesticMSTRecords(raw, src.TailLen, src.Market, src.Exchange)
		for _, rec := range recs {
			idx.add(rec)
			total++
		}
	}

	for _, src := range overseasMasterSources {
		raw, err := downloadMasterPayload(ctx, httpClient, src.URL)
		if err != nil {
			continue
		}
		recs := parseOverseasCODRecords(raw, src.Market, src.Exchange, src.Country)
		for _, rec := range recs {
			idx.add(rec)
			total++
		}
	}

	if total == 0 {
		return idx, 0, fmt.Errorf("symbol bootstrap failed: no master symbols loaded")
	}

	return idx, total, nil
}

func (idx *masterSymbolsIndex) add(rec MasterSymbol) {
	sym := strings.ToUpper(strings.TrimSpace(rec.Symbol))
	if sym == "" {
		return
	}
	mkt := canonicalMasterMarket(rec.Market)
	if mkt == "" {
		return
	}

	key := mkt + "|" + sym
	if _, ok := idx.byMarketSymbol[key]; !ok {
		rec.Market = mkt
		idx.byMarketSymbol[key] = rec
	}
	if _, ok := idx.anyBySym[sym]; !ok {
		idx.anyBySym[sym] = rec
	}

	if isDomesticMarket(mkt) {
		if _, ok := idx.domesticBySym[sym]; !ok {
			idx.domesticBySym[sym] = rec
		}
	}
	if isUSMarket(mkt) {
		if _, ok := idx.usBySym[sym]; !ok {
			idx.usBySym[sym] = rec
		}
	}
}

func lookupMasterSymbolFromIndex(idx masterSymbolsIndex, market, symbol string) (MasterSymbol, bool) {
	sym := normalizeLookupSymbol(symbol)
	if sym == "" {
		return MasterSymbol{}, false
	}

	mkt := canonicalLookupMarket(market)
	if mkt != "" {
		if rec, ok := idx.byMarketSymbol[mkt+"|"+sym]; ok {
			return rec, true
		}
	}

	if mkt == "" || isDomesticMarket(mkt) || mkt == "KRX" {
		if rec, ok := idx.domesticBySym[sym]; ok {
			return rec, true
		}
	}

	if isUSMarket(mkt) || mkt == "US" {
		if rec, ok := idx.usBySym[sym]; ok {
			return rec, true
		}
	}

	rec, ok := idx.anyBySym[sym]
	return rec, ok
}

func downloadMasterPayload(ctx context.Context, httpClient *http.Client, rawURL string) ([]byte, error) {
	u := strings.TrimSpace(rawURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download %s: HTTP %d", u, resp.StatusCode)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(strings.ToLower(u), ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if f.FileInfo().IsDir() {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(rc)
			_ = rc.Close()
			if err == nil {
				return data, nil
			}
		}
		return nil, fmt.Errorf("zip %s has no readable file", u)
	}

	return payload, nil
}

func parseDomesticMSTRecords(raw []byte, tailLen int, market, exchange string) []MasterSymbol {
	recs := make([]MasterSymbol, 0, 1024)
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	for scanner.Scan() {
		line := bytes.TrimRight(scanner.Bytes(), "\r\n")
		if len(line) <= tailLen || len(line) < 21 {
			continue
		}
		part1 := line[:len(line)-tailLen]
		if len(part1) < 21 {
			continue
		}

		shortCode := strings.TrimSpace(string(part1[:9]))
		stdCode := strings.TrimSpace(string(part1[9:21]))
		name := decodeCP949(bytes.TrimSpace(part1[21:]))

		symbol := extractDomesticSixDigitCode(shortCode, stdCode)
		if symbol == "" {
			continue
		}
		recs = append(recs, MasterSymbol{
			Symbol:          symbol,
			Market:          market,
			Name:            name,
			Exchange:        exchange,
			Currency:        "KRW",
			Country:         "KR",
			ProductType:     "stock",
			ProductTypeCode: "300",
			IsListed:        true,
		})
	}

	return recs
}

func parseOverseasCODRecords(raw []byte, market, defaultExchange, defaultCountry string) []MasterSymbol {
	recs := make([]MasterSymbol, 0, 2048)
	decoded := decodeCP949(raw)
	scanner := bufio.NewScanner(strings.NewReader(decoded))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 10 {
			continue
		}

		symbol := strings.ToUpper(strings.TrimSpace(fields[4]))
		if symbol == "" || strings.EqualFold(symbol, "symbol") {
			continue
		}

		exchangeCode := strings.ToUpper(strings.TrimSpace(fields[2]))
		if exchangeCode == "" {
			exchangeCode = defaultExchange
		}
		name := strings.TrimSpace(fields[6])
		nameEn := strings.TrimSpace(fields[7])
		securityType := strings.TrimSpace(fields[8])
		currency := strings.ToUpper(strings.TrimSpace(fields[9]))
		if currency == "" {
			currency = "USD"
		}

		productType := "stock"
		switch securityType {
		case "1":
			productType = "index"
		case "3":
			productType = "etf"
		case "4":
			productType = "warrant"
		}

		recs = append(recs, MasterSymbol{
			Symbol:          symbol,
			Market:          market,
			Name:            name,
			NameEn:          nameEn,
			Exchange:        exchangeCode,
			Currency:        currency,
			Country:         defaultCountry,
			ProductType:     productType,
			ProductTypeCode: productTypeCodeFromOverseasMarket(market),
			IsListed:        true,
		})
	}

	return recs
}

func decodeCP949(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	r := transform.NewReader(bytes.NewReader(raw), korean.EUCKR.NewDecoder())
	decoded, err := io.ReadAll(r)
	if err != nil {
		return strings.TrimSpace(string(raw))
	}
	return strings.TrimSpace(string(decoded))
}

func extractDomesticSixDigitCode(candidates ...string) string {
	for _, c := range candidates {
		code := findSixDigits(c)
		if code != "" {
			return code
		}
	}
	return ""
}

func findSixDigits(s string) string {
	run := 0
	start := -1
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= '0' && ch <= '9' {
			if run == 0 {
				start = i
			}
			run++
			if run >= 6 {
				return s[start : start+6]
			}
			continue
		}
		run = 0
		start = -1
	}
	return ""
}

func normalizeLookupSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	if s == "" {
		return ""
	}
	if len(s) > 1 && s[0] == 'A' {
		if c := findSixDigits(s); c != "" {
			return c
		}
	}
	if c := findSixDigits(s); c != "" && len(s) <= 9 {
		return c
	}
	return s
}

func canonicalLookupMarket(market string) string {
	m := strings.ToUpper(strings.TrimSpace(market))
	switch m {
	case "", "KRX", "KOSPI", "KOSDAQ", "KONEX", "KNX", "NXT":
		if m == "KNX" {
			return "KONEX"
		}
		return m
	case "US", "US-NASDAQ", "NASDAQ", "NAS", "NASD":
		if m == "US" {
			return "US"
		}
		return "US-NASDAQ"
	case "US-NYSE", "NYSE", "NYS":
		return "US-NYSE"
	case "US-AMEX", "AMEX", "AMS":
		return "US-AMEX"
	case "HK", "HKS", "SEHK":
		return "HK"
	case "JP", "JAPAN", "TKSE", "TSE":
		return "JP"
	case "SH", "SHAA", "SHS":
		return "SH"
	case "SZ", "SZAA", "SZS":
		return "SZ"
	case "HNX", "HASE":
		return "HNX"
	case "HSX", "VNSE":
		return "HSX"
	default:
		return m
	}
}

func canonicalMasterMarket(market string) string {
	return canonicalLookupMarket(market)
}

func isDomesticMarket(market string) bool {
	switch market {
	case "KRX", "KOSPI", "KOSDAQ", "KONEX", "NXT":
		return true
	default:
		return false
	}
}

func isUSMarket(market string) bool {
	switch market {
	case "US", "US-NASDAQ", "US-NYSE", "US-AMEX":
		return true
	default:
		return false
	}
}

func productTypeCodeFromOverseasMarket(market string) string {
	switch canonicalLookupMarket(market) {
	case "US-NASDAQ":
		return "512"
	case "US-NYSE":
		return "513"
	case "US-AMEX":
		return "529"
	case "JP":
		return "515"
	case "HK":
		return "501"
	case "SH":
		return "551"
	case "SZ":
		return "552"
	case "HNX":
		return "507"
	case "HSX":
		return "508"
	default:
		return ""
	}
}
