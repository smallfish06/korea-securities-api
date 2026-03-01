package adapter

import (
	"context"
	"errors"
	"net/http"
	"reflect"
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

func TestMarshalMap_ArrayWrappedAsItems(t *testing.T) {
	t.Parallel()

	out, err := marshalMap([]string{"a", "b"}, nil)
	if err != nil {
		t.Fatalf("marshalMap error: %v", err)
	}

	raw, ok := out["items"]
	if !ok {
		t.Fatalf("items key missing: %#v", out)
	}
	items, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("items type = %T, want []interface{}", raw)
	}
	want := []interface{}{"a", "b"}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("items = %#v, want %#v", items, want)
	}
}

func TestApplyDocumentedCustomDefaults_TicScope(t *testing.T) {
	t.Parallel()

	payload := map[string]interface{}{}
	applyDocumentedDefaults("ka50079", payload)
	if payload["tic_scope"] != "1" {
		t.Fatalf("tic_scope = %#v, want \"1\"", payload["tic_scope"])
	}

	payload2 := map[string]interface{}{"tic_scope": "5"}
	applyDocumentedDefaults("ka50080", payload2)
	if payload2["tic_scope"] != "5" {
		t.Fatalf("tic_scope = %#v, want \"5\"", payload2["tic_scope"])
	}
}
