package kis

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	kisspecs "github.com/smallfish06/krsec/pkg/kis/specs"
)

func TestCallDocumentedEndpoint_GET(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-stock/v1/quotations/chk-holiday" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "CTCA0903R" {
			t.Fatalf("unexpected tr_id: %s", got)
		}
		if got := r.URL.Query().Get("BASS_DT"); got != "20260302" {
			t.Fatalf("unexpected query BASS_DT: %s", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok","output":{"is_holiday":"N"}}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp := kisspecs.NewDocumentedEndpointResponse("/uapi/domestic-stock/v1/quotations/chk-holiday")
	if resp == nil {
		t.Fatal("expected documented response type")
	}
	err := client.CallDocumentedEndpointInto(context.Background(), http.MethodGet, "/uapi/domestic-stock/v1/quotations/chk-holiday", "CTCA0903R", map[string]string{
		"BASS_DT": "20260302",
	}, resp)
	if err != nil {
		t.Fatalf("CallDocumentedEndpointInto GET returned error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatalf("expected success response")
	}
}

func TestCallDocumentedEndpoint_POST(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/uapi/domestic-bond/v1/trading/buy" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("tr_id"); got != "CTSC3008U" {
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
		if body["PDNO"] != "KR1234567890" {
			t.Fatalf("unexpected PDNO: %s", body["PDNO"])
		}
		if body["ORD_QTY"] != "1" {
			t.Fatalf("unexpected ORD_QTY: %s", body["ORD_QTY"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"0","msg_cd":"00000","msg1":"ok"}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp := kisspecs.NewDocumentedEndpointResponse("/uapi/domestic-bond/v1/trading/buy")
	if resp == nil {
		t.Fatal("expected documented response type")
	}
	err := client.CallDocumentedEndpointInto(context.Background(), http.MethodPost, "/uapi/domestic-bond/v1/trading/buy", "CTSC3008U", map[string]string{
		"CANO":         "12345678",
		"ACNT_PRDT_CD": "01",
		"PDNO":         "KR1234567890",
		"ORD_QTY":      "1",
	}, resp)
	if err != nil {
		t.Fatalf("CallDocumentedEndpointInto POST returned error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatalf("expected success response")
	}
}

func TestCallDocumentedEndpoint_POST_APIError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rt_cd":"1","msg_cd":"EGW00123","msg1":"invalid request"}`))
	}))
	defer ts.Close()

	client := newAuthedTestClient(ts.URL)
	resp := kisspecs.NewDocumentedEndpointResponse("/uapi/domestic-bond/v1/trading/buy")
	if resp == nil {
		t.Fatal("expected documented response type")
	}
	err := client.CallDocumentedEndpointInto(context.Background(), http.MethodPost, "/uapi/domestic-bond/v1/trading/buy", "CTSC3008U", map[string]string{
		"PDNO": "KR1234567890",
	}, resp)
	if err == nil {
		t.Fatal("expected API error, got nil")
	}
}
