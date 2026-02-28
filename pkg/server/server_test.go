package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

type fakeBroker struct{}

func (fakeBroker) Name() string { return "FAKE" }

func (fakeBroker) Authenticate(context.Context, broker.Credentials) (*broker.Token, error) {
	return &broker.Token{AccessToken: "t", TokenType: "Bearer", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (fakeBroker) GetQuote(context.Context, string, string) (*broker.Quote, error) {
	return &broker.Quote{}, nil
}

func (fakeBroker) GetOHLCV(context.Context, string, string, broker.OHLCVOpts) ([]broker.OHLCV, error) {
	return []broker.OHLCV{}, nil
}

func (fakeBroker) GetBalance(context.Context, string) (*broker.Balance, error) {
	return &broker.Balance{}, nil
}

func (fakeBroker) GetPositions(context.Context, string) ([]broker.Position, error) {
	return []broker.Position{}, nil
}

func (fakeBroker) PlaceOrder(context.Context, broker.OrderRequest) (*broker.OrderResult, error) {
	return &broker.OrderResult{}, nil
}

func (fakeBroker) CancelOrder(context.Context, string) error { return nil }

func (fakeBroker) ModifyOrder(context.Context, string, broker.ModifyOrderRequest) (*broker.OrderResult, error) {
	return &broker.OrderResult{}, nil
}

func TestNew_HealthAndAccounts(t *testing.T) {
	t.Parallel()

	s := New(Options{
		Host: "127.0.0.1",
		Port: 18080,
		Accounts: []Account{
			{ID: "12345678-01", Name: "main", Broker: "custom"},
		},
		Brokers: map[string]broker.Broker{
			"12345678-01": fakeBroker{},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	s.App().Mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected health status: %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rr = httptest.NewRecorder()
	s.App().Mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected accounts status: %d", rr.Code)
	}
	body := rr.Body.Bytes()

	var got struct {
		OK   bool                 `json:"ok"`
		Data []broker.AccountInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal accounts response: %v", err)
	}
	if !got.OK {
		t.Fatalf("expected ok=true, got false body=%s", string(body))
	}
	if len(got.Data) != 1 {
		t.Fatalf("expected one account, got %d", len(got.Data))
	}
	if got.Data[0].ID != "12345678-01" || got.Data[0].Broker != "custom" {
		t.Fatalf("unexpected account row: %+v", got.Data[0])
	}
}

func TestNew_OpenAPIEndpoints(t *testing.T) {
	t.Parallel()

	s := New(Options{
		Host: "127.0.0.1",
		Port: 18081,
		Accounts: []Account{
			{ID: "12345678-01", Name: "main", Broker: "custom"},
		},
		Brokers: map[string]broker.Broker{
			"12345678-01": fakeBroker{},
		},
	})

	// In non-Run tests, OpenAPI routes are explicitly registered on the mux.
	s.App().RegisterOpenAPIRoutes(s.App())

	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil)
	rr := httptest.NewRecorder()
	s.App().Mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected openapi spec status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"openapi":"3.1.0"`) {
		t.Fatalf("unexpected openapi spec body: %s", rr.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	rr = httptest.NewRecorder()
	s.App().Mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected swagger ui status: %d body=%s", rr.Code, rr.Body.String())
	}
}
