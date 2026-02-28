package kis

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInquireStockBasicInfo(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/quotations/search-stock-info" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "CTPF1002R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		q := r.URL.Query()
		if q.Get("PRDT_TYPE_CD") != "300" || q.Get("PDNO") != "005930" {
			t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"pdno":"005930","prdt_type_cd":"300","mket_id_cd":"STK","scty_grp_id_cd":"ST","excg_dvsn_cd":"02","prdt_name":"삼성전자","prdt_eng_name":"SAMSUNG ELECTRONICS","tr_stop_yn":"N"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireStockBasicInfo(context.Background(), "005930", "300")
	if err != nil {
		t.Fatalf("InquireStockBasicInfo returned error: %v", err)
	}
	if resp.Output.PdNo != "005930" {
		t.Fatalf("unexpected pdno: %s", resp.Output.PdNo)
	}
	if resp.Output.PrdtName != "삼성전자" {
		t.Fatalf("unexpected prdt_name: %s", resp.Output.PrdtName)
	}
}

func TestInquireProductBasicInfo(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/quotations/search-info" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "CTPF1604R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		q := r.URL.Query()
		if q.Get("PRDT_TYPE_CD") != "300" || q.Get("PDNO") != "000660" {
			t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
		}

		// Validate request has no body for GET
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var asMap map[string]interface{}
		if len(b) > 0 {
			if err := json.Unmarshal(b, &asMap); err != nil {
				t.Fatalf("decode body: %v", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"pdno":"000660","prdt_type_cd":"300","prdt_name":"SK하이닉스","prdt_eng_name":"SK HYNIX"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireProductBasicInfo(context.Background(), "000660", "300")
	if err != nil {
		t.Fatalf("InquireProductBasicInfo returned error: %v", err)
	}
	if resp.Output.PdNo != "000660" {
		t.Fatalf("unexpected pdno: %s", resp.Output.PdNo)
	}
}

func TestInquireOverseasProductBasicInfo(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/overseas-price/v1/quotations/search-info" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "CTPF1702R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		q := r.URL.Query()
		if q.Get("PRDT_TYPE_CD") != "512" || q.Get("PDNO") != "AAPL" {
			t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"std_pdno":"AAPL","prdt_name":"Apple Inc.","prdt_eng_name":"Apple Inc.","natn_name":"미국","ovrs_excg_cd":"NASD","tr_crcy_cd":"USD","ovrs_stck_dvsn_cd":"01","lstg_yn":"Y","lstg_dt":"19801212","ovrs_stck_tr_stop_dvsn_cd":"01"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquireOverseasProductBasicInfo(context.Background(), "AAPL", "512")
	if err != nil {
		t.Fatalf("InquireOverseasProductBasicInfo returned error: %v", err)
	}
	if resp.Output.StdPdNo != "AAPL" {
		t.Fatalf("unexpected std_pdno: %s", resp.Output.StdPdNo)
	}
}
