package kis

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOrderOverseas_USBuy(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/overseas-stock/v1/trading/order" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTT1002U" {
			t.Fatalf("unexpected tr_id: %s", got)
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var body map[string]string
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["OVRS_EXCG_CD"] != "NASD" {
			t.Fatalf("unexpected OVRS_EXCG_CD: %s", body["OVRS_EXCG_CD"])
		}
		if body["PDNO"] != "AAPL" {
			t.Fatalf("unexpected PDNO: %s", body["PDNO"])
		}
		if body["ORD_DVSN"] != "00" {
			t.Fatalf("unexpected ORD_DVSN: %s", body["ORD_DVSN"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"ODNO":"09000123","ORD_TMD":"223001"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.OrderOverseas(context.Background(), "12345678", "01", "NASD", "AAPL", 1, 223.45, "buy", "00")
	if err != nil {
		t.Fatalf("OrderOverseas returned error: %v", err)
	}
	if resp.Output.OrdNo != "09000123" {
		t.Fatalf("unexpected order number: %s", resp.Output.OrdNo)
	}
}

func TestOrderOverseasRvseCncl_USCancel(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/overseas-stock/v1/trading/order-rvsecncl" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTT1004U" {
			t.Fatalf("unexpected tr_id: %s", got)
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var body map[string]string
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["ORGN_ODNO"] != "09000123" {
			t.Fatalf("unexpected ORGN_ODNO: %s", body["ORGN_ODNO"])
		}
		if body["RVSE_CNCL_DVSN_CD"] != "02" {
			t.Fatalf("unexpected RVSE_CNCL_DVSN_CD: %s", body["RVSE_CNCL_DVSN_CD"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"ODNO":"09000123","ORD_TMD":"223101"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.OrderOverseasRvseCncl(context.Background(), "12345678", "01", "NASD", "AAPL", "09000123", "02", 1, 0)
	if err != nil {
		t.Fatalf("OrderOverseasRvseCncl returned error: %v", err)
	}
	if resp.Output.OrdNo != "09000123" {
		t.Fatalf("unexpected order number: %s", resp.Output.OrdNo)
	}
}

func TestNormalizeOverseasExchangeCode(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"us":      "NASD",
		"nas":     "NASD",
		"nasd":    "NASD",
		"us-nyse": "NYSE",
		"amex":    "AMEX",
		"sehk":    "SEHK",
		"unknown": "UNKNOWN",
	}

	for in, want := range tests {
		got := normalizeOverseasExchangeCode(in)
		if got != want {
			t.Fatalf("normalizeOverseasExchangeCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInquireOverseasPrice_UsesOverseasPricePath(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/overseas-price/v1/quotations/price" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "HHDFS00000300" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		if got := r.URL.Query().Get("EXCD"); got != "NAS" {
			t.Fatalf("unexpected EXCD: %s", got)
		}
		if got := r.URL.Query().Get("SYMB"); got != "AAPL" {
			t.Fatalf("unexpected SYMB: %s", got)
		}
		if got := r.URL.Query().Get("AUTH"); got != "" {
			t.Fatalf("unexpected AUTH: %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"rt_cd":"0",
			"msg_cd":"00000",
			"msg1":"ok",
			"output":{
				"rsym":"DNASAAPL",
				"symb_desc":"APPLE INC",
				"last":"220.12",
				"open":"219.00",
				"high":"221.00",
				"low":"218.10",
				"prdy_vrss":"1.12",
				"prdy_vrss_sign":"2",
				"t_xvol":"1234567"
			}
		}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireOverseasPrice(context.Background(), "NASDAQ", "AAPL")
	if err != nil {
		t.Fatalf("InquireOverseasPrice returned error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatalf("expected success response, got: %+v", resp)
	}
	if got := strings.TrimSpace(resp.Output.Last); got != "220.12" {
		t.Fatalf("unexpected last price: %s", got)
	}
}
