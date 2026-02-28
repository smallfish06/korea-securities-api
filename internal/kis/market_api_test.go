package kis

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInquireAskingPriceExpCcn(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/quotations/inquire-asking-price-exp-ccn" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("tr_id"); got != "FHKST01010200" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		if q := r.URL.Query().Get("FID_INPUT_ISCD"); q != "005930" {
			t.Fatalf("unexpected symbol: %s", q)
		}
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output1":{"askp1":"70100"},"output2":{"exp_prc":"70050"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireAskingPriceExpCcn(context.Background(), "J", "005930")
	if err != nil {
		t.Fatalf("InquireAskingPriceExpCcn returned error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatalf("expected success: %+v", resp)
	}
}

func TestInquireTimeItemConclusion(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/quotations/inquire-time-itemconclusion" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("tr_id"); got != "FHPST01060000" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		if q := r.URL.Query().Get("FID_INPUT_HOUR_1"); q != "090000" {
			t.Fatalf("unexpected hour: %s", q)
		}
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output1":{"stck_prpr":"70000"},"output2":[{"cnqn":"10"}]}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	if _, err := client.InquireTimeItemConclusion(context.Background(), "J", "005930", "090000"); err != nil {
		t.Fatalf("InquireTimeItemConclusion returned error: %v", err)
	}
}

func TestInquireOverseasDailyPrice(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/overseas-price/v1/quotations/dailyprice" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("tr_id"); got != "HHDFS76240000" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		if q := r.URL.Query().Get("SYMB"); q != "AAPL" {
			t.Fatalf("unexpected symbol: %s", q)
		}
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output1":{"last":"230.10"},"output2":[{"xymd":"20260227"}]}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	if _, err := client.InquireOverseasDailyPrice(context.Background(), "", "NAS", "AAPL", "0", "", "0"); err != nil {
		t.Fatalf("InquireOverseasDailyPrice returned error: %v", err)
	}
}

func TestInquirePossibleOrder(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/trading/inquire-psbl-order" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("tr_id"); got != "TTTC8908R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		if q := r.URL.Query().Get("CANO"); q != "12345678" {
			t.Fatalf("unexpected cano: %s", q)
		}
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"max_buy_qty":"10"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	if _, err := client.InquirePossibleOrder(context.Background(), "12345678", "01", "005930", "70000", "00", "N", "N"); err != nil {
		t.Fatalf("InquirePossibleOrder returned error: %v", err)
	}
}

func TestStockDailyPriceResponseRowsPrefersOutputThenOutput1(t *testing.T) {
	t.Parallel()

	r := StockDailyPriceResponse{Output: []StockDailyPriceOutput{{StckBsopDate: "20260227"}}}
	if got := len(r.Rows()); got != 1 {
		t.Fatalf("rows len = %d, want 1", got)
	}

	r = StockDailyPriceResponse{Output1: []StockDailyPriceOutput{{StckBsopDate: "20260226"}}}
	if got := len(r.Rows()); got != 1 {
		t.Fatalf("rows len = %d, want 1", got)
	}
}

func TestStockDailyPriceResponse_ParsesOutput2(t *testing.T) {
	t.Parallel()

	var r StockDailyPriceResponse
	payload := []byte(`{
		"rt_cd":"0",
		"msg_cd":"00000",
		"msg1":"ok",
		"output1":[{"stck_bsop_date":"20260227","stck_clpr":"70000"}],
		"output2":{"last_stck_bsop_date":"20260227","next_stck_bsop_date":"20260228"}
	}`)
	if err := json.Unmarshal(payload, &r); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got := len(r.Rows()); got != 1 {
		t.Fatalf("rows len = %d, want 1", got)
	}
	if r.Output2["last_stck_bsop_date"] != "20260227" {
		t.Fatalf("unexpected output2 parse: %+v", r.Output2)
	}
}
