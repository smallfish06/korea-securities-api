package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/smallfish06/krsec/pkg/config"
)

type proxyKiwoomBroker struct {
	proxyStubBroker
	called    bool
	gotMethod string
	gotPath   string
	gotAPIID  string
	gotFields map[string]string
	resp      map[string]interface{}
	err       error
}

func (b *proxyKiwoomBroker) CallEndpoint(
	_ context.Context,
	method string,
	path string,
	apiID string,
	fields map[string]string,
) (map[string]interface{}, error) {
	b.called = true
	b.gotMethod = method
	b.gotPath = path
	b.gotAPIID = apiID
	b.gotFields = fields
	return b.resp, b.err
}

func TestHandleKiwoomProxy_DefaultRouteAndFirstKiwoomAccount(t *testing.T) {
	t.Parallel()

	kiwoomBroker := &proxyKiwoomBroker{
		proxyStubBroker: proxyStubBroker{name: "KIWOOM"},
		resp:            map[string]interface{}{"return_code": 0, "return_msg": "ok"},
	}
	kisBroker := &proxyStubBroker{name: "KIS"}

	s := newOrderTestServer(
		map[string]broker.Broker{
			"kis-acc":    kisBroker,
			"kiwoom-acc": kiwoomBroker,
		},
		[]config.AccountConfig{
			{AccountID: "kis-acc", Broker: "kis"},
			{AccountID: "kiwoom-acc", Broker: "kiwoom"},
		},
	)

	body := []byte(`{"api_id":"ka10001","params":{"stk_cd":"005930"}}`)
	req := httptest.NewRequest(http.MethodPost, "/kiwoom/dostk/stkinfo", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeResponse(t, rr)
	if !resp.OK {
		t.Fatalf("expected ok=true")
	}
	if resp.Broker != "KIWOOM" {
		t.Fatalf("broker = %q, want KIWOOM", resp.Broker)
	}
	if !kiwoomBroker.called {
		t.Fatalf("expected Kiwoom broker to be called")
	}
	if kiwoomBroker.gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", kiwoomBroker.gotMethod)
	}
	if kiwoomBroker.gotPath != "/api/dostk/stkinfo" {
		t.Fatalf("path = %q", kiwoomBroker.gotPath)
	}
	if kiwoomBroker.gotAPIID != "ka10001" {
		t.Fatalf("api_id = %q, want ka10001", kiwoomBroker.gotAPIID)
	}
	if got := kiwoomBroker.gotFields["stk_cd"]; got != "005930" {
		t.Fatalf("params stk_cd = %q, want 005930", got)
	}
}

func TestHandleKiwoomProxy_MissingAPIID(t *testing.T) {
	t.Parallel()

	kiwoomBroker := &proxyKiwoomBroker{proxyStubBroker: proxyStubBroker{name: "KIWOOM"}}
	s := newOrderTestServer(
		map[string]broker.Broker{"kiwoom-acc": kiwoomBroker},
		[]config.AccountConfig{{AccountID: "kiwoom-acc", Broker: "kiwoom"}},
	)

	body := []byte(`{"params":{"stk_cd":"005930"}}`)
	req := httptest.NewRequest(http.MethodPost, "/kiwoom/dostk/stkinfo", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleKiwoomProxy_InvalidAccount(t *testing.T) {
	t.Parallel()

	kiwoomBroker := &proxyKiwoomBroker{proxyStubBroker: proxyStubBroker{name: "KIWOOM"}}
	s := newOrderTestServer(
		map[string]broker.Broker{"kiwoom-acc": kiwoomBroker},
		[]config.AccountConfig{{AccountID: "kiwoom-acc", Broker: "kiwoom"}},
	)

	body := []byte(`{"account_id":"missing","api_id":"ka10001","params":{"stk_cd":"005930"}}`)
	req := httptest.NewRequest(http.MethodPost, "/kiwoom/dostk/stkinfo", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestHandleKiwoomProxy_NonKiwoomAccountRejected(t *testing.T) {
	t.Parallel()

	kisBroker := &proxyStubBroker{name: "KIS"}
	s := newOrderTestServer(
		map[string]broker.Broker{"kis-acc": kisBroker},
		[]config.AccountConfig{{AccountID: "kis-acc", Broker: "kis"}},
	)

	body := []byte(`{"account_id":"kis-acc","api_id":"ka10001","params":{"stk_cd":"005930"}}`)
	req := httptest.NewRequest(http.MethodPost, "/kiwoom/dostk/stkinfo", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}
