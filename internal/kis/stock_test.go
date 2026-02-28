package kis

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newAuthedTestClient(baseURL string) *Client {
	c := NewClient(false)
	c.baseURL = baseURL
	c.SetCredentials("app", "secret")
	c.setToken("token", time.Now().Add(time.Hour))
	return c
}

func TestOrderCash(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/trading/order-cash" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTC0802U" {
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

		if body["CANO"] != "12345678" {
			t.Fatalf("unexpected CANO: %s", body["CANO"])
		}
		if body["ACNT_PRDT_CD"] != "01" {
			t.Fatalf("unexpected ACNT_PRDT_CD: %s", body["ACNT_PRDT_CD"])
		}
		if body["PDNO"] != "005930" {
			t.Fatalf("unexpected PDNO: %s", body["PDNO"])
		}
		if body["ORD_DVSN"] != "00" {
			t.Fatalf("unexpected ORD_DVSN: %s", body["ORD_DVSN"])
		}
		if body["EXCG_ID_DVSN_CD"] != "KRX" {
			t.Fatalf("unexpected EXCG_ID_DVSN_CD: %s", body["EXCG_ID_DVSN_CD"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"KRX_FWDG_ORD_ORGNO":"06010","ODNO":"000123","ORD_TMD":"101010"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.OrderCash(context.Background(), "12345678", "01", "005930", "limit", 10, 70000, "buy", "KRX")
	if err != nil {
		t.Fatalf("OrderCash returned error: %v", err)
	}
	if resp.Output.OrdNo != "000123" {
		t.Fatalf("unexpected order number: %s", resp.Output.OrdNo)
	}
	if resp.Output.KrxFwdOrdOrgno != "06010" {
		t.Fatalf("unexpected order orgno: %s", resp.Output.KrxFwdOrdOrgno)
	}
}

func TestOrderRvseCncl(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/trading/order-rvsecncl" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTC0803U" {
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

		if body["KRX_FWDG_ORD_ORGNO"] != "06010" {
			t.Fatalf("unexpected KRX_FWDG_ORD_ORGNO: %s", body["KRX_FWDG_ORD_ORGNO"])
		}
		if body["ORGN_ODNO"] != "000123" {
			t.Fatalf("unexpected ORGN_ODNO: %s", body["ORGN_ODNO"])
		}
		if body["RVSE_CNCL_DVSN_CD"] != "02" {
			t.Fatalf("unexpected RVSE_CNCL_DVSN_CD: %s", body["RVSE_CNCL_DVSN_CD"])
		}
		if body["QTY_ALL_ORD_YN"] != "Y" {
			t.Fatalf("unexpected QTY_ALL_ORD_YN: %s", body["QTY_ALL_ORD_YN"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"KRX_FWDG_ORD_ORGNO":"06010","ODNO":"000123","ORD_TMD":"101011"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.OrderRvseCncl(
		context.Background(),
		"12345678",
		"01",
		"06010",
		"000123",
		"00",
		"02",
		10,
		70000,
		true,
		"KRX",
	)
	if err != nil {
		t.Fatalf("OrderRvseCncl returned error: %v", err)
	}
	if resp.Output.OrdNo != "000123" {
		t.Fatalf("unexpected order number: %s", resp.Output.OrdNo)
	}
}

func TestInquirePossibleRvseCncl(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/trading/inquire-psbl-rvsecncl" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "TTTC8036R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		q := r.URL.Query()
		if q.Get("CANO") != "12345678" || q.Get("ACNT_PRDT_CD") != "01" {
			t.Fatalf("unexpected query params: %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":[{"odno":"000123","orgn_odno":"000123","ord_gno_brno":"06010","ord_dvsn":"00","ord_qty":"10","ord_unpr":"70000","psbl_qty":"10","excg_id_dvsn_cd":"KRX"}],"ctx_area_fk100":"","ctx_area_nk100":""}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp, err := client.InquirePossibleRvseCncl(context.Background(), "12345678", "01")
	if err != nil {
		t.Fatalf("InquirePossibleRvseCncl returned error: %v", err)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("unexpected output length: %d", len(resp.Output))
	}
	if resp.Output[0].OrdGnoBrno != "06010" {
		t.Fatalf("unexpected ord_gno_brno: %s", resp.Output[0].OrdGnoBrno)
	}
}
