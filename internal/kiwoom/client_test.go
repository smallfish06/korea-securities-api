package kiwoom

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

type memoryTokenManager struct {
	token     string
	expiresAt time.Time
	hasToken  bool

	setCalls  int
	waitCalls int
}

func (m *memoryTokenManager) GetToken(string) (string, time.Time, bool) {
	if m.hasToken {
		return m.token, m.expiresAt, true
	}
	return "", time.Time{}, false
}

func (m *memoryTokenManager) SetToken(_ string, token string, expiresAt time.Time) error {
	m.token = token
	m.expiresAt = expiresAt
	m.hasToken = true
	m.setCalls++
	return nil
}

func (m *memoryTokenManager) WaitForAuth(string) {
	m.waitCalls++
}

func TestLookupAPISpec_KnownTR(t *testing.T) {
	spec, ok, err := LookupAPISpec("kt10000")
	if err != nil {
		t.Fatalf("LookupAPISpec error: %v", err)
	}
	if !ok {
		t.Fatal("expected kt10000 spec to exist")
	}
	if spec.Path != "/api/dostk/ordr" {
		t.Fatalf("path = %q, want /api/dostk/ordr", spec.Path)
	}
}

func TestClientGetDomesticQuote_UsesAuthAndAPIIDHeader(t *testing.T) {
	var gotAuth string
	var gotAPIID string
	var gotSymbol string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"expires_dt":  "20991231235959",
				"token_type":  "bearer",
				"token":       "test-token",
				"return_code": 0,
				"return_msg":  "ok",
			})
		case "/api/dostk/stkinfo":
			gotAuth = r.Header.Get("Authorization")
			gotAPIID = r.Header.Get("api-id")
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			gotSymbol = body["stk_cd"]
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"stk_cd":      "005930",
				"cur_prc":     "70000",
				"open_pric":   "69500",
				"high_pric":   "70500",
				"low_pric":    "69000",
				"pred_pre":    "500",
				"flu_rt":      "0.72",
				"trde_qty":    "12345",
				"base_pric":   "69500",
				"upl_pric":    "91000",
				"lst_pric":    "48000",
				"return_code": 0,
				"return_msg":  "ok",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	tm := &memoryTokenManager{}
	c := NewClientWithTokenManager(false, tm)
	c.SetBaseURL(ts.URL)

	if _, err := c.Authenticate(context.Background(), broker.Credentials{AppKey: "k", AppSecret: "s"}); err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}

	quote, err := c.GetDomesticQuote(context.Background(), "005930")
	if err != nil {
		t.Fatalf("GetDomesticQuote error: %v", err)
	}

	if tm.setCalls != 1 {
		t.Fatalf("SetToken calls = %d, want 1", tm.setCalls)
	}
	if !strings.HasPrefix(gotAuth, "Bearer test-token") {
		t.Fatalf("Authorization header = %q", gotAuth)
	}
	if gotAPIID != "ka10001" {
		t.Fatalf("api-id header = %q, want ka10001", gotAPIID)
	}
	if gotSymbol != "005930" {
		t.Fatalf("stk_cd = %q, want 005930", gotSymbol)
	}
	if quote.Price != 70000 || quote.Change != 500 || quote.ChangeRate != 0.72 {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestClientGetDomesticQuote_ReturnCodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"expires_dt":  "20991231235959",
				"token_type":  "bearer",
				"token":       "test-token",
				"return_code": 0,
				"return_msg":  "ok",
			})
		case "/api/dostk/stkinfo":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"return_code": -301,
				"return_msg":  "invalid symbol",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := NewClientWithTokenManager(false, &memoryTokenManager{})
	c.SetBaseURL(ts.URL)
	c.SetCredentials("k", "s")

	_, err := c.GetDomesticQuote(context.Background(), "BAD")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid symbol") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(err, broker.ErrInvalidSymbol) {
		t.Fatalf("expected ErrInvalidSymbol mapping, got: %v", err)
	}
}

func TestClientPlaceStockOrder_SellUsesSellTR(t *testing.T) {
	var gotAPIID string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"expires_dt":  "20991231235959",
				"token_type":  "bearer",
				"token":       "test-token",
				"return_code": 0,
				"return_msg":  "ok",
			})
		case "/api/dostk/ordr":
			gotAPIID = r.Header.Get("api-id")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ord_no":      "0001234",
				"return_code": 0,
				"return_msg":  "ok",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := NewClientWithTokenManager(false, &memoryTokenManager{})
	c.SetBaseURL(ts.URL)
	if _, err := c.Authenticate(context.Background(), broker.Credentials{AppKey: "k", AppSecret: "s"}); err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}

	_, err := c.PlaceStockOrder(context.Background(), PlaceStockOrderRequest{
		Side:       StockOrderSideSell,
		Exchange:   "KRX",
		Symbol:     "005930",
		Quantity:   1,
		OrderPrice: "70000",
		TradeType:  "0",
	})
	if err != nil {
		t.Fatalf("PlaceStockOrder error: %v", err)
	}
	if gotAPIID != "kt10001" {
		t.Fatalf("api-id header = %q, want kt10001", gotAPIID)
	}
}

func TestAuthenticate_InvalidCredentialsMapped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"return_code": 401,
			"return_msg":  "appkey invalid",
		})
	}))
	defer ts.Close()

	c := NewClientWithTokenManager(false, &memoryTokenManager{})
	c.SetBaseURL(ts.URL)

	_, err := c.Authenticate(context.Background(), broker.Credentials{AppKey: "bad", AppSecret: "bad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, broker.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials mapping, got: %v", err)
	}
}
