package adapter

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/smallfish06/krsec/internal/kiwoom"
	"github.com/smallfish06/krsec/pkg/broker"
)

func TestCallEndpoint_RequiresAPIID(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodPost, kiwoom.PathStockInfo, "", map[string]string{"stk_cd": "005930"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCallEndpoint_UnsupportedPath(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodPost, "/api/unknown/path", kiwoom.APIIDDomesticQuote, map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCallEndpoint_UnsupportedPathMethodValidation(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodGet, "/api/unknown/path", "zz99999", map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCallEndpoint_UnsupportedAPIIDOnKnownPath(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodPost, kiwoom.PathStockInfo, "zz99999", map[string]string{"stk_cd": "005930"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCallEndpoint_DocumentedRouteRequiresClient(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodPost, kiwoom.PathStockInfo, "ka10002", map[string]string{"stk_cd": "005930"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCallEndpoint_MethodValidation(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodGet, kiwoom.PathStockInfo, kiwoom.APIIDDomesticQuote, map[string]string{"stk_cd": "005930"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidOrderRequest) {
		t.Fatalf("expected ErrInvalidOrderRequest, got %v", err)
	}
}

func TestCallEndpoint_ValidRouteMissingSymbol(t *testing.T) {
	t.Parallel()
	a := &Adapter{}

	_, err := a.CallEndpoint(context.Background(), http.MethodPost, kiwoom.PathStockInfo, kiwoom.APIIDDomesticQuote, map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, broker.ErrInvalidSymbol) {
		t.Fatalf("expected ErrInvalidSymbol, got %v", err)
	}
}
