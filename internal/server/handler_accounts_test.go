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

func TestHandleGetBalance_AmbiguousAccountReturnsBadRequest(t *testing.T) {
	t.Parallel()

	first := newMockBroker(t, "KIS-1")
	second := newMockBroker(t, "KIS-2")
	s := newOrderTestServer(
		map[string]broker.Broker{
			"12345678-01": first,
			"12345678-02": second,
		},
		[]config.AccountConfig{
			{AccountID: "12345678-01"},
			{AccountID: "12345678-02"},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/12345678/balance", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeResponse(t, rr)
	if resp.Error != ambiguousAccountIDError {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}
