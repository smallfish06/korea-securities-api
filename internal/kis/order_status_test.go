package kis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInquireDailyCcld(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/trading/inquire-daily-ccld" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTC0081R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		q := r.URL.Query()
		if q.Get("CANO") != "12345678" || q.Get("ACNT_PRDT_CD") != "01" {
			t.Fatalf("unexpected account query: %s", r.URL.RawQuery)
		}
		if q.Get("ODNO") != "000123" {
			t.Fatalf("unexpected ODNO: %s", q.Get("ODNO"))
		}
		if q.Get("EXCG_ID_DVSN_CD") != "KRX" {
			t.Fatalf("unexpected EXCG_ID_DVSN_CD: %s", q.Get("EXCG_ID_DVSN_CD"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output1":[{"ord_dt":"20260228","ord_gno_brno":"06010","odno":"000123","orgn_odno":"000123","pdno":"005930","ord_qty":"10","tot_ccld_qty":"10","rmn_qty":"0","cncl_yn":"N","rjct_qty":"0"}],"ctx_area_fk100":"","ctx_area_nk100":""}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireDailyCcld(context.Background(), "12345678", "01", "20260228", "20260228", "06010", "000123", "KRX")
	if err != nil {
		t.Fatalf("InquireDailyCcld returned error: %v", err)
	}
	if len(resp.Output1) != 1 {
		t.Fatalf("unexpected output length: %d", len(resp.Output1))
	}
	if resp.Output1[0].ODNo != "000123" {
		t.Fatalf("unexpected ODNO: %s", resp.Output1[0].ODNo)
	}
}

func TestInquireOverseasCcnl(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/overseas-stock/v1/trading/inquire-ccnl" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTS3035R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		q := r.URL.Query()
		if q.Get("CANO") != "12345678" || q.Get("ACNT_PRDT_CD") != "01" {
			t.Fatalf("unexpected account query: %s", r.URL.RawQuery)
		}
		if q.Get("OVRS_EXCG_CD") != "NASD" {
			t.Fatalf("unexpected OVRS_EXCG_CD: %s", q.Get("OVRS_EXCG_CD"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","ctx_area_fk200":"","ctx_area_nk200":"","output":[{"ord_dt":"20260228","odno":"09000123","orgn_odno":"09000123","pdno":"AAPL","ft_ord_qty":"5","ft_ccld_qty":"5","nccs_qty":"0","prcs_stat_name":"완료"}]}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireOverseasCcnl(context.Background(), "12345678", "01", "20260228", "20260228", "NASD")
	if err != nil {
		t.Fatalf("InquireOverseasCcnl returned error: %v", err)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("unexpected output length: %d", len(resp.Output))
	}
	if resp.Output[0].ODNo != "09000123" {
		t.Fatalf("unexpected ODNO: %s", resp.Output[0].ODNo)
	}
}

func TestInquireOverseasCcnl_PaginatesUntilCursorEnd(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch strings.TrimSpace(q.Get("CTX_AREA_NK200")) {
		case "":
			_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","ctx_area_fk200":"fk1","ctx_area_nk200":"nk1","output":[{"ord_dt":"20260228","odno":"09000123","orgn_odno":"09000123","pdno":"AAPL","ft_ord_qty":"5","ft_ccld_qty":"5","nccs_qty":"0","prcs_stat_name":"완료"}]}`))
		case "nk1":
			_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","ctx_area_fk200":"","ctx_area_nk200":"","output":[{"ord_dt":"20260228","odno":"09000124","orgn_odno":"09000124","pdno":"AAPL","ft_ord_qty":"3","ft_ccld_qty":"3","nccs_qty":"0","prcs_stat_name":"완료"}]}`))
		default:
			t.Fatalf("unexpected cursor: %s", q.Get("CTX_AREA_NK200"))
		}
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireOverseasCcnl(context.Background(), "12345678", "01", "20260228", "20260228", "NASD")
	if err != nil {
		t.Fatalf("InquireOverseasCcnl returned error: %v", err)
	}
	if len(resp.Output) != 2 {
		t.Fatalf("unexpected output length: %d", len(resp.Output))
	}
}
