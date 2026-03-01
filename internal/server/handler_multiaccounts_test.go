package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/smallfish06/krsec/pkg/config"
)

func TestHandleAccountsSummary_ReturnsServiceUnavailableWhenAllBalancesFail(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetBalance", mock.Anything, "acc1").Return((*broker.Balance)(nil), errors.New("upstream unavailable")).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/summary", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeResponse(t, rr)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
	if resp.Error != "failed to retrieve balances from all accounts" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}
