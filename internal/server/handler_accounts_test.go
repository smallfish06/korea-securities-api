package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/smallfish06/krsec/pkg/config"
)

func TestHandleGetBalance_UnknownAccountReturnsNotFound(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/missing/balance", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
	if resp.Error != "account not found" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}

func TestHandleGetPositions_UnknownAccountReturnsNotFound(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/missing/positions", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeResponse(t, rr)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
	if resp.Error != "account not found" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}
