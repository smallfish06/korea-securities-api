package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/smallfish06/krsec/pkg/broker"
	apiserver "github.com/smallfish06/krsec/pkg/server"
)

type demoBroker struct{}

func (d *demoBroker) Name() string { return "DEMO" }

func (d *demoBroker) Authenticate(_ context.Context, creds broker.Credentials) (*broker.Token, error) {
	if strings.TrimSpace(creds.AppKey) == "" || strings.TrimSpace(creds.AppSecret) == "" {
		return nil, broker.ErrInvalidCredentials
	}
	return &broker.Token{
		AccessToken: "demo-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

func (d *demoBroker) GetQuote(_ context.Context, market, symbol string) (*broker.Quote, error) {
	if strings.TrimSpace(market) == "" {
		return nil, broker.ErrInvalidMarket
	}
	if strings.TrimSpace(symbol) == "" {
		return nil, broker.ErrInvalidSymbol
	}
	return &broker.Quote{
		Symbol:    strings.ToUpper(symbol),
		Market:    strings.ToUpper(market),
		Price:     70000,
		Open:      69500,
		High:      71000,
		Low:       69000,
		Close:     69800,
		Volume:    123456,
		Timestamp: time.Now().UTC(),
	}, nil
}

func (d *demoBroker) GetOHLCV(_ context.Context, market, symbol string, _ broker.OHLCVOpts) ([]broker.OHLCV, error) {
	if strings.TrimSpace(market) == "" {
		return nil, broker.ErrInvalidMarket
	}
	if strings.TrimSpace(symbol) == "" {
		return nil, broker.ErrInvalidSymbol
	}
	return []broker.OHLCV{{
		Timestamp: time.Now().UTC().Add(-24 * time.Hour),
		Open:      69000,
		High:      70500,
		Low:       68800,
		Close:     70000,
		Volume:    98765,
	}}, nil
}

func (d *demoBroker) GetBalance(_ context.Context, accountID string) (*broker.Balance, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, broker.ErrInvalidOrderRequest
	}
	return &broker.Balance{
		AccountID:   accountID,
		Cash:        1000000,
		TotalAssets: 1500000,
		BuyingPower: 1200000,
		ProfitLoss:  50000,
	}, nil
}

func (d *demoBroker) GetPositions(_ context.Context, _ string) ([]broker.Position, error) {
	return []broker.Position{{
		Symbol:       "005930",
		Name:         "삼성전자",
		Market:       "KRX",
		AssetType:    broker.AssetStock,
		Quantity:     5,
		AvgPrice:     68000,
		CurrentPrice: 70000,
		ProfitLoss:   10000,
	}}, nil
}

func (d *demoBroker) PlaceOrder(_ context.Context, req broker.OrderRequest) (*broker.OrderResult, error) {
	if req.Quantity <= 0 {
		return nil, broker.ErrInvalidOrderRequest
	}
	return &broker.OrderResult{
		OrderID:   "DEMO-1",
		Status:    broker.OrderStatusFilled,
		Message:   "accepted",
		Timestamp: time.Now().UTC(),
	}, nil
}

func (d *demoBroker) CancelOrder(_ context.Context, _ string) error { return nil }

func (d *demoBroker) ModifyOrder(_ context.Context, _ string, _ broker.ModifyOrderRequest) (*broker.OrderResult, error) {
	return &broker.OrderResult{
		OrderID:   "DEMO-1",
		Status:    broker.OrderStatusPending,
		Message:   "modified",
		Timestamp: time.Now().UTC(),
	}, nil
}

func main() {
	const (
		host      = "127.0.0.1"
		port      = 19090
		accountID = "demo-acc-1"
	)

	srv := apiserver.New(apiserver.Options{
		Host: host,
		Port: port,
		Accounts: []apiserver.Account{{
			ID:     accountID,
			Name:   "Demo Account",
			Broker: "demo",
			Credentials: broker.Credentials{
				AppKey:    "demo-key",
				AppSecret: "demo-secret",
			},
		}},
		Brokers: map[string]broker.Broker{accountID: &demoBroker{}},
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run()
	}()

	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	if err := waitForHealth(baseURL+"/health", 5*time.Second); err != nil {
		log.Fatalf("server health check failed: %v", err)
	}

	mustPrintJSON("GET /health", baseURL+"/health")
	mustPrintJSON("GET /quotes/KRX/005930", baseURL+"/quotes/KRX/005930")
	mustPrintJSON("GET /accounts/demo-acc-1/balance", baseURL+"/accounts/demo-acc-1/balance")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.App().Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}

	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("run: %v", err)
	}
}

func waitForHealth(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", url)
}

func mustPrintJSON(label, url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("%s request failed: %v", label, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%s read body failed: %v", label, err)
	}

	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		log.Printf("%s (%d): %s", label, resp.StatusCode, strings.TrimSpace(string(body)))
		return
	}

	pretty, err := json.MarshalIndent(decoded, "", "  ")
	if err != nil {
		log.Fatalf("%s pretty print failed: %v", label, err)
	}
	log.Printf("%s (%d):\n%s", label, resp.StatusCode, string(pretty))
}
