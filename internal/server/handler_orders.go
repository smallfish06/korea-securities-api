package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"

	"github.com/smallfish06/krsec/pkg/broker"
)

type orderGetter interface {
	GetOrder(ctx context.Context, orderID string) (*broker.OrderResult, error)
}

type orderFillsGetter interface {
	GetOrderFills(ctx context.Context, orderID string) ([]broker.OrderFill, error)
}

// handleGetOrder handles GET /accounts/{account_id}/orders/{order_id}
func (s *Server) handleGetOrder(c fuego.ContextNoBody) (Response, error) {
	accountID := c.PathParam("account_id")
	orderID := c.PathParam("order_id")
	brk, ok := s.getBrokerStrict(accountID)
	if !ok {
		return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
	}

	getter, ok := brk.(orderGetter)
	if !ok {
		return respond(c, http.StatusNotImplemented, Response{
			OK:    false,
			Error: "order status lookup not supported by broker",
		})
	}

	result, err := getter.GetOrder(c.Context(), orderID)
	if err == nil {
		return respond(c, http.StatusOK, Response{
			OK:     true,
			Data:   result,
			Broker: brk.Name(),
		})
	}
	if errors.Is(err, broker.ErrOrderNotFound) {
		return respond(c, http.StatusNotFound, Response{
			OK:    false,
			Error: "order not found",
		})
	}

	return respond(c, statusFromBrokerError(err, http.StatusInternalServerError), Response{
		OK:    false,
		Error: err.Error(),
	})
}

// handleGetOrderFills handles GET /accounts/{account_id}/orders/{order_id}/fills
func (s *Server) handleGetOrderFills(c fuego.ContextNoBody) (Response, error) {
	accountID := c.PathParam("account_id")
	orderID := c.PathParam("order_id")
	brk, ok := s.getBrokerStrict(accountID)
	if !ok {
		return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
	}

	getter, ok := brk.(orderFillsGetter)
	if !ok {
		return respond(c, http.StatusNotImplemented, Response{
			OK:    false,
			Error: "order fills lookup not supported by broker",
		})
	}

	fills, err := getter.GetOrderFills(c.Context(), orderID)
	if err == nil {
		return respond(c, http.StatusOK, Response{
			OK:     true,
			Data:   fills,
			Broker: brk.Name(),
		})
	}
	if errors.Is(err, broker.ErrOrderNotFound) {
		return respond(c, http.StatusNotFound, Response{
			OK:    false,
			Error: "order not found",
		})
	}

	return respond(c, statusFromBrokerError(err, http.StatusInternalServerError), Response{
		OK:    false,
		Error: err.Error(),
	})
}

// handlePlaceOrder handles POST /accounts/{account_id}/orders
func (s *Server) handlePlaceOrder(c fuego.ContextWithBody[broker.OrderRequest]) (Response, error) {
	accountID := c.PathParam("account_id")

	brk, ok := s.getBrokerStrict(accountID)
	if !ok {
		return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
	}

	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "invalid request body",
		})
	}

	if req.AccountID != "" && !sameAccountID(req.AccountID, accountID) {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "account_id in body does not match path",
		})
	}
	req.AccountID = accountID

	result, err := brk.PlaceOrder(c.Context(), req)
	if err != nil {
		status := statusFromBrokerError(err, http.StatusInternalServerError)
		return respond(c, status, Response{
			OK:    false,
			Error: err.Error(),
		})
	}

	return respond(c, http.StatusOK, Response{
		OK:     true,
		Data:   result,
		Broker: brk.Name(),
	})
}

// handleCancelOrder handles DELETE /accounts/{account_id}/orders/{order_id}
func (s *Server) handleCancelOrder(c fuego.ContextNoBody) (Response, error) {
	accountID := c.PathParam("account_id")
	orderID := c.PathParam("order_id")
	brk, ok := s.getBrokerStrict(accountID)
	if !ok {
		return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
	}

	err := brk.CancelOrder(c.Context(), orderID)
	if err == nil {
		return respond(c, http.StatusOK, Response{
			OK:     true,
			Broker: brk.Name(),
		})
	}

	if errors.Is(err, broker.ErrOrderNotFound) {
		return respond(c, http.StatusNotFound, Response{
			OK:    false,
			Error: "order not found",
		})
	}

	return respond(c, statusFromBrokerError(err, http.StatusInternalServerError), Response{
		OK:    false,
		Error: err.Error(),
	})
}

// handleModifyOrder handles PUT /accounts/{account_id}/orders/{order_id}
func (s *Server) handleModifyOrder(c fuego.ContextWithBody[broker.ModifyOrderRequest]) (Response, error) {
	accountID := c.PathParam("account_id")
	orderID := c.PathParam("order_id")
	brk, ok := s.getBrokerStrict(accountID)
	if !ok {
		return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
	}

	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "invalid request body",
		})
	}

	result, err := brk.ModifyOrder(c.Context(), orderID, req)
	if err == nil {
		return respond(c, http.StatusOK, Response{
			OK:     true,
			Data:   result,
			Broker: brk.Name(),
		})
	}

	if errors.Is(err, broker.ErrOrderNotFound) {
		return respond(c, http.StatusNotFound, Response{
			OK:    false,
			Error: "order not found",
		})
	}

	return respond(c, statusFromBrokerError(err, http.StatusInternalServerError), Response{
		OK:    false,
		Error: err.Error(),
	})
}

func (s *Server) orderBrokerCandidates(accountID string) []broker.Broker {
	out := make([]broker.Broker, 0, len(s.brokers)+1)
	seen := make(map[broker.Broker]struct{})

	if accountID != "" {
		if brk, ok := s.getBrokerStrict(accountID); ok {
			out = append(out, brk)
			seen[brk] = struct{}{}
		}
	}

	for _, brk := range s.brokers {
		if _, ok := seen[brk]; ok {
			continue
		}
		out = append(out, brk)
		seen[brk] = struct{}{}
	}

	return out
}

func (s *Server) getBrokerStrict(accountID string) (broker.Broker, bool) {
	if brk, ok := s.brokers[accountID]; ok {
		return brk, true
	}
	for key, brk := range s.brokers {
		if strings.HasPrefix(key, accountID+"-") || strings.HasPrefix(accountID, key+"-") || strings.TrimSuffix(key, "-01") == strings.TrimSuffix(accountID, "-01") {
			return brk, true
		}
	}
	return nil, false
}

func sameAccountID(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == b {
		return true
	}
	if strings.TrimSuffix(a, "-01") == strings.TrimSuffix(b, "-01") {
		return true
	}
	if strings.HasPrefix(a, b+"-") || strings.HasPrefix(b, a+"-") {
		return true
	}
	return false
}
