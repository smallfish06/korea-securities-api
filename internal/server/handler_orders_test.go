package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	testifymock "github.com/stretchr/testify/mock"

	servermock "github.com/smallfish06/krsec/internal/server/mock"
	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/smallfish06/krsec/pkg/config"
)

func newOrderTestServer(brokers map[string]broker.Broker, accounts []config.AccountConfig) *Server {
	s := newBaseServer(&config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 18080,
		},
		Accounts: accounts,
	})
	s.brokers = brokers
	s.accounts = accounts
	return s
}

func performFiberRequest(t *testing.T, s *Server, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	rr := httptest.NewRecorder()
	s.router.Mux.ServeHTTP(rr, req)
	return rr
}

func newMockBroker(t *testing.T, name string) *servermock.MockBrokerWithExtras {
	t.Helper()

	b := servermock.NewMockBrokerWithExtras(t)
	b.On("Name").Return(name).Maybe()
	return b
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) Response {
	t.Helper()
	var resp Response
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestHandlePlaceOrder_BodyAccountMismatchReturnsBadRequest(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b, "acc2": b},
		[]config.AccountConfig{{AccountID: "acc1"}, {AccountID: "acc2"}},
	)

	body := []byte(`{"account_id":"acc2","symbol":"005930","market":"KRX","side":"buy","type":"limit","quantity":1,"price":70000}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts/acc1/orders", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	resp := decodeResponse(t, rr)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
}

func TestHandlePlaceOrder_UsesBrokerDomainShape(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("PlaceOrder", testifymock.Anything, testifymock.Anything).Return(&broker.OrderResult{
		OrderID:   "000123",
		Status:    broker.OrderStatusPending,
		Message:   "accepted",
		Timestamp: time.Now(),
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	body := []byte(`{"symbol":"005930","market":"KRX","side":"buy","type":"limit","quantity":1,"price":70000}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts/acc1/orders", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if bytes.Contains(rr.Body.Bytes(), []byte(`"ODNO"`)) {
		t.Fatalf("unexpected raw KIS field leaked in response: %s", rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"order_id"`)) {
		t.Fatalf("expected broker domain field order_id in response: %s", rr.Body.String())
	}
}

func TestHandleGetOrder_ReturnsDomainOrderResult(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetOrder", testifymock.Anything, "000123").Return(&broker.OrderResult{
		OrderID:   "000123",
		Status:    broker.OrderStatusFilled,
		Message:   "done",
		Timestamp: time.Now(),
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/acc1/orders/000123", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if bytes.Contains(rr.Body.Bytes(), []byte(`"ODNO"`)) {
		t.Fatalf("unexpected raw KIS field leaked in response: %s", rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"order_id"`)) {
		t.Fatalf("expected broker domain field order_id in response: %s", rr.Body.String())
	}
}

func TestHandleGetOrder_Returns404WhenOrderNotFound(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetOrder", testifymock.Anything, "000123").Return((*broker.OrderResult)(nil), broker.ErrOrderNotFound).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/acc1/orders/000123", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleCancelOrder_UsesPathAccountBrokerOnly(t *testing.T) {
	t.Parallel()

	firstCalled := false
	secondCalled := false

	b1 := newMockBroker(t, "KIS-1")
	b1.On("CancelOrder", testifymock.Anything, "000123").Run(func(args testifymock.Arguments) {
		firstCalled = true
	}).Return(nil).Once()
	b2 := newMockBroker(t, "KIS-2")
	b2.On("CancelOrder", testifymock.Anything, "000123").Run(func(args testifymock.Arguments) {
		secondCalled = true
	}).Return(nil).Maybe()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b1, "acc2": b2},
		[]config.AccountConfig{{AccountID: "acc1"}, {AccountID: "acc2"}},
	)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/acc1/orders/000123", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !firstCalled {
		t.Fatalf("expected first broker to be attempted")
	}
	if secondCalled {
		t.Fatalf("expected second broker not to be attempted")
	}
	resp := decodeResponse(t, rr)
	if !resp.OK {
		t.Fatalf("expected ok=true")
	}
}

func TestHandleCancelOrder_Returns404WhenOrderNotFound(t *testing.T) {
	t.Parallel()

	b1 := newMockBroker(t, "KIS-1")
	b1.On("CancelOrder", testifymock.Anything, "000123").Return(broker.ErrOrderNotFound).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b1},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/acc1/orders/000123", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	resp := decodeResponse(t, rr)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
}

func TestHandleModifyOrder_Returns404WhenOrderNotFound(t *testing.T) {
	t.Parallel()

	b1 := newMockBroker(t, "KIS-1")
	b1.On("ModifyOrder", testifymock.Anything, "000123", testifymock.Anything).Return((*broker.OrderResult)(nil), broker.ErrOrderNotFound).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b1},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	body := []byte(`{"quantity":1,"price":70000}`)
	req := httptest.NewRequest(http.MethodPut, "/accounts/acc1/orders/000123", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
	resp := decodeResponse(t, rr)
	if resp.OK {
		t.Fatalf("expected ok=false")
	}
}

func TestHandleGetOrderFills_ReturnsDomainFills(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetOrderFills", testifymock.Anything, "000123").Return([]broker.OrderFill{
		{
			OrderID:  "000123",
			Symbol:   "005930",
			Market:   "KRX",
			Quantity: 10,
			Price:    70000,
		},
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/acc1/orders/000123/fills", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if bytes.Contains(rr.Body.Bytes(), []byte(`"ODNO"`)) {
		t.Fatalf("unexpected raw KIS field leaked in response: %s", rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"order_id"`)) {
		t.Fatalf("expected broker domain field order_id in response: %s", rr.Body.String())
	}
}

func TestHandleGetOrderFills_Returns404WhenOrderNotFound(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetOrderFills", testifymock.Anything, "000123").Return(([]broker.OrderFill)(nil), broker.ErrOrderNotFound).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/accounts/acc1/orders/000123/fills", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
