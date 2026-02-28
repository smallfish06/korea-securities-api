package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/smallfish06/krsec/pkg/broker"
	"github.com/smallfish06/krsec/pkg/config"
)

func TestHandleGetInstrument_ReturnsDomainInstrument(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetInstrument", mock.Anything, "KRX", "005930").Return(&broker.Instrument{
		Symbol:          "005930",
		Market:          "KRX",
		Name:            "삼성전자",
		NameEn:          "SAMSUNG ELECTRONICS",
		Exchange:        "KRX",
		Currency:        "KRW",
		AssetType:       broker.AssetStock,
		ProductType:     "stock",
		ProductTypeCode: "300",
		IsListed:        true,
	}, nil).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/instruments/KRX/005930", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte(`"symbol"`)) {
		t.Fatalf("expected normalized symbol field in response: %s", rr.Body.String())
	}
	if bytes.Contains(rr.Body.Bytes(), []byte(`"pdno"`)) {
		t.Fatalf("unexpected raw broker field leaked in response: %s", rr.Body.String())
	}
}

func TestHandleGetInstrument_Returns404WhenNotFound(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	b.On("GetInstrument", mock.Anything, "KRX", "INVALID").Return((*broker.Instrument)(nil), broker.ErrInstrumentNotFound).Once()

	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/instruments/KRX/INVALID", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleGetInstrument_Returns404WhenAccountNotFound(t *testing.T) {
	t.Parallel()

	b := newMockBroker(t, "KIS")
	s := newOrderTestServer(
		map[string]broker.Broker{"acc1": b},
		[]config.AccountConfig{{AccountID: "acc1"}},
	)

	req := httptest.NewRequest(http.MethodGet, "/instruments/KRX/005930?account_id=missing", nil)
	rr := performFiberRequest(t, s, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
