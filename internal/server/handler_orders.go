package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/smallfish06/kr-broker-api/pkg/broker"
)

type orderGetter interface {
	GetOrder(ctx context.Context, orderID string) (*broker.OrderResult, error)
}

type orderFillsGetter interface {
	GetOrderFills(ctx context.Context, orderID string) ([]broker.OrderFill, error)
}

// handleGetOrder handles GET /orders/{order_id}
func (s *Server) handleGetOrder(c fuego.ContextNoBody) (Response, error) {
	orderID := c.PathParam("order_id")
	accountID := c.QueryParam("account_id")
	if accountID != "" {
		if _, ok := s.getBrokerStrict(accountID); !ok {
			return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
		}
	}

	candidates := s.orderBrokerCandidates(accountID)
	if len(candidates) == 0 {
		return respond(c, http.StatusServiceUnavailable, Response{OK: false, Error: "no broker available"})
	}

	var firstErr error
	supported := false
	for _, brk := range candidates {
		getter, ok := brk.(orderGetter)
		if !ok {
			continue
		}
		supported = true

		result, err := getter.GetOrder(c.Context(), orderID)
		if err == nil {
			return respond(c, http.StatusOK, Response{
				OK:     true,
				Data:   result,
				Broker: brk.Name(),
			})
		}
		if errors.Is(err, broker.ErrOrderNotFound) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return respond(c, http.StatusInternalServerError, Response{
			OK:    false,
			Error: firstErr.Error(),
		})
	}
	if !supported {
		return respond(c, http.StatusNotImplemented, Response{
			OK:    false,
			Error: "order status lookup not supported by broker",
		})
	}

	return respond(c, http.StatusNotFound, Response{
		OK:    false,
		Error: "order not found",
	})
}

// handleGetOrderFills handles GET /orders/{order_id}/fills
func (s *Server) handleGetOrderFills(c fuego.ContextNoBody) (Response, error) {
	orderID := c.PathParam("order_id")
	accountID := c.QueryParam("account_id")
	if accountID != "" {
		if _, ok := s.getBrokerStrict(accountID); !ok {
			return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
		}
	}

	candidates := s.orderBrokerCandidates(accountID)
	if len(candidates) == 0 {
		return respond(c, http.StatusServiceUnavailable, Response{OK: false, Error: "no broker available"})
	}

	var firstErr error
	supported := false
	for _, brk := range candidates {
		getter, ok := brk.(orderFillsGetter)
		if !ok {
			continue
		}
		supported = true

		fills, err := getter.GetOrderFills(c.Context(), orderID)
		if err == nil {
			return respond(c, http.StatusOK, Response{
				OK:     true,
				Data:   fills,
				Broker: brk.Name(),
			})
		}
		if errors.Is(err, broker.ErrOrderNotFound) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return respond(c, http.StatusInternalServerError, Response{
			OK:    false,
			Error: firstErr.Error(),
		})
	}
	if !supported {
		return respond(c, http.StatusNotImplemented, Response{
			OK:    false,
			Error: "order fills lookup not supported by broker",
		})
	}

	return respond(c, http.StatusNotFound, Response{
		OK:    false,
		Error: "order not found",
	})
}

// handlePlaceOrder handles POST /orders
func (s *Server) handlePlaceOrder(c fuego.ContextWithBody[broker.OrderRequest]) (Response, error) {
	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "invalid request body",
		})
	}

	if req.AccountID == "" && len(s.accounts) > 1 {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "account_id is required in multi-account mode",
		})
	}

	var brk broker.Broker
	if req.AccountID != "" {
		var ok bool
		brk, ok = s.getBrokerStrict(req.AccountID)
		if !ok {
			return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
		}
	} else {
		brk = s.getFirstBroker()
	}
	if brk == nil {
		return respond(c, http.StatusServiceUnavailable, Response{OK: false, Error: "no broker available"})
	}

	result, err := brk.PlaceOrder(c.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, broker.ErrInvalidOrderRequest) {
			status = http.StatusBadRequest
		}
		if errors.Is(err, broker.ErrOrderNotFound) {
			status = http.StatusNotFound
		}
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

// handleCancelOrder handles DELETE /orders/{order_id}
func (s *Server) handleCancelOrder(c fuego.ContextNoBody) (Response, error) {
	orderID := c.PathParam("order_id")
	accountID := c.QueryParam("account_id")
	if accountID != "" {
		if _, ok := s.getBrokerStrict(accountID); !ok {
			return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
		}
	}
	candidates := s.orderBrokerCandidates(accountID)
	if len(candidates) == 0 {
		return respond(c, http.StatusServiceUnavailable, Response{OK: false, Error: "no broker available"})
	}

	var firstErr error
	for _, brk := range candidates {
		err := brk.CancelOrder(c.Context(), orderID)
		if err == nil {
			return respond(c, http.StatusOK, Response{
				OK:     true,
				Broker: brk.Name(),
			})
		}
		if errors.Is(err, broker.ErrOrderNotFound) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		status := http.StatusInternalServerError
		if errors.Is(firstErr, broker.ErrInvalidOrderRequest) {
			status = http.StatusBadRequest
		}
		return respond(c, status, Response{
			OK:    false,
			Error: firstErr.Error(),
		})
	}

	return respond(c, http.StatusNotFound, Response{
		OK:    false,
		Error: "order not found",
	})
}

// handleModifyOrder handles PUT /orders/{order_id}
func (s *Server) handleModifyOrder(c fuego.ContextWithBody[broker.ModifyOrderRequest]) (Response, error) {
	orderID := c.PathParam("order_id")

	req, err := c.Body()
	if err != nil {
		return respond(c, http.StatusBadRequest, Response{
			OK:    false,
			Error: "invalid request body",
		})
	}

	accountID := c.QueryParam("account_id")
	if accountID != "" {
		if _, ok := s.getBrokerStrict(accountID); !ok {
			return respond(c, http.StatusNotFound, Response{OK: false, Error: "account not found"})
		}
	}
	candidates := s.orderBrokerCandidates(accountID)
	if len(candidates) == 0 {
		return respond(c, http.StatusServiceUnavailable, Response{OK: false, Error: "no broker available"})
	}

	var firstErr error
	for _, brk := range candidates {
		result, err := brk.ModifyOrder(c.Context(), orderID, req)
		if err == nil {
			return respond(c, http.StatusOK, Response{
				OK:     true,
				Data:   result,
				Broker: brk.Name(),
			})
		}
		if errors.Is(err, broker.ErrOrderNotFound) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		status := http.StatusInternalServerError
		if errors.Is(firstErr, broker.ErrInvalidOrderRequest) {
			status = http.StatusBadRequest
		}
		return respond(c, status, Response{
			OK:    false,
			Error: firstErr.Error(),
		})
	}

	return respond(c, http.StatusNotFound, Response{
		OK:    false,
		Error: "order not found",
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
