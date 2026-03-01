package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/smallfish06/krsec/pkg/config"
)

func TestHandleAuthToken_RespectsRequestedBroker(t *testing.T) {
	t.Parallel()

	kis := newMockBroker(t, "KIS")
	kiwoom := newMockBroker(t, "KIWOOM")
	kis.On("Authenticate", mock.Anything, broker.Credentials{
		AppKey:    "k",
		AppSecret: "s",
	}).Return(&broker.Token{
		AccessToken: "kis-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{
			"kiwoom-acc": kiwoom,
			"kis-acc":    kis,
		},
		[]config.AccountConfig{
			{AccountID: "kiwoom-acc", Broker: "kiwoom", Sandbox: true},
			{AccountID: "kis-acc", Broker: "kis", Sandbox: true},
		},
	)

	body := []byte(`{"broker":"kis","credentials":{"app_key":"k","app_secret":"s"},"sandbox":true}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeResponse(t, rr)
	if !resp.OK {
		t.Fatalf("expected ok=true")
	}
	if resp.Broker != "KIS" {
		t.Fatalf("broker = %q, want KIS", resp.Broker)
	}
	kiwoom.AssertNotCalled(t, "Authenticate", mock.Anything, mock.Anything)
}

func TestHandleAuthToken_RejectsUnsupportedBroker(t *testing.T) {
	t.Parallel()

	kis := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"kis-acc": kis},
		[]config.AccountConfig{{AccountID: "kis-acc", Broker: "kis"}},
	)

	body := []byte(`{"broker":"future","credentials":{"app_key":"k","app_secret":"s"}}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeResponse(t, rr)
	if resp.Error != "unsupported broker" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	kis.AssertNotCalled(t, "Authenticate", mock.Anything, mock.Anything)
}

func TestHandleAuthToken_SelectsSandboxMatchFirst(t *testing.T) {
	t.Parallel()

	sandboxBroker := newMockBroker(t, "KIS")
	prodBroker := newMockBroker(t, "KIS")

	sandboxBroker.On("Authenticate", mock.Anything, mock.Anything).Return(&broker.Token{
		AccessToken: "sandbox-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{
			"kis-prod":    prodBroker,
			"kis-sandbox": sandboxBroker,
		},
		[]config.AccountConfig{
			{AccountID: "kis-prod", Broker: "kis", Sandbox: false},
			{AccountID: "kis-sandbox", Broker: "kis", Sandbox: true},
		},
	)

	body := []byte(`{"broker":"kis","credentials":{"app_key":"k","app_secret":"s"},"sandbox":true}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(body))
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	prodBroker.AssertNotCalled(t, "Authenticate", mock.Anything, mock.Anything)
}
