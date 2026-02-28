package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smallfish06/krsec/internal/config"
	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/stretchr/testify/mock"
)

func TestHandleGetOHLCV_ParsesOptions(t *testing.T) {
	t.Parallel()

	var captured broker.OHLCVOpts
	b := newMockBroker(t, "KIS")
	b.On("GetOHLCV", mock.Anything, "KRX", "005930", mock.Anything).Run(func(args mock.Arguments) {
		captured = args.Get(3).(broker.OHLCVOpts)
	}).Return([]broker.OHLCV{
		{Timestamp: time.Now(), Open: 1, High: 1, Low: 1, Close: 1, Volume: 1},
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/quotes/KRX/005930/ohlcv?interval=1w&from=2026-01-01&to=2026-01-31&limit=50", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if captured.Interval != "1w" || captured.Limit != 50 {
		t.Fatalf("unexpected opts: %+v", captured)
	}
	if captured.From.Format("2006-01-02") != "2026-01-01" {
		t.Fatalf("unexpected from: %s", captured.From.Format("2006-01-02"))
	}
	if captured.To.Format("2006-01-02") != "2026-01-31" {
		t.Fatalf("unexpected to: %s", captured.To.Format("2006-01-02"))
	}
}

func TestHandleGetQuote_InvalidSymbolReturnsBadRequest(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetQuote", mock.Anything, "KRX", "BAD").Return((*broker.Quote)(nil), broker.ErrInvalidSymbol).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/quotes/KRX/BAD", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestStatusFromBrokerError_DefaultAndTyped(t *testing.T) {
	t.Parallel()

	if got := statusFromBrokerError(broker.ErrInvalidMarket, http.StatusInternalServerError); got != http.StatusBadRequest {
		t.Fatalf("invalid market status = %d, want 400", got)
	}
	if got := statusFromBrokerError(broker.ErrOrderNotFound, http.StatusInternalServerError); got != http.StatusNotFound {
		t.Fatalf("order not found status = %d, want 404", got)
	}
	if got := statusFromBrokerError(errors.New("unknown"), http.StatusInternalServerError); got != http.StatusInternalServerError {
		t.Fatalf("unknown status = %d, want 500", got)
	}
}

func TestHandleGetOHLCV_InvalidLimitReturnsBadRequest(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/quotes/KRX/005930/ohlcv?limit=abc", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleGetOHLCV_InvalidIntervalReturnsBadRequest(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/quotes/KRX/005930/ohlcv?interval=5m", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
