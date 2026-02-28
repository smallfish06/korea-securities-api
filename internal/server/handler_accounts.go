package server

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// handleGetBalance handles GET /accounts/{account_id}/balance
func (s *Server) handleGetBalance(c fuego.ContextNoBody) (Response, error) {
	accountID := c.PathParam("account_id")

	brk := s.getBroker(accountID)
	if brk == nil {
		return respond(c, http.StatusNotFound, Response{
			OK:    false,
			Error: "account not found",
		})
	}

	balance, err := brk.GetBalance(c.Context(), accountID)
	if err != nil {
		return respond(c, http.StatusInternalServerError, Response{
			OK:    false,
			Error: err.Error(),
		})
	}

	return respond(c, http.StatusOK, Response{
		OK:     true,
		Data:   balance,
		Broker: brk.Name(),
	})
}

// handleGetPositions handles GET /accounts/{account_id}/positions
func (s *Server) handleGetPositions(c fuego.ContextNoBody) (Response, error) {
	accountID := c.PathParam("account_id")

	brk := s.getBroker(accountID)
	if brk == nil {
		return respond(c, http.StatusNotFound, Response{
			OK:    false,
			Error: "account not found",
		})
	}

	positions, err := brk.GetPositions(c.Context(), accountID)
	if err != nil {
		return respond(c, http.StatusInternalServerError, Response{
			OK:    false,
			Error: err.Error(),
		})
	}

	// Check if grouping is requested
	if c.QueryParam("group") == "true" {
		positions = groupPositions(positions)
	}

	return respond(c, http.StatusOK, Response{
		OK:     true,
		Data:   positions,
		Broker: brk.Name(),
	})
}

// groupPositions groups positions by symbol and asset type
func groupPositions(positions []broker.Position) []broker.Position {
	grouped := make(map[string]*broker.Position)
	totalAmt := make(map[string]float64) // 매입금액 합계 추적

	for _, p := range positions {
		key := p.Symbol + "|" + string(p.AssetType)
		if existing, ok := grouped[key]; ok {
			totalAmt[key] += p.AvgPrice * float64(p.Quantity)
			existing.Quantity += p.Quantity
			existing.AvgPrice = totalAmt[key] / float64(existing.Quantity)
			existing.ProfitLoss += p.ProfitLoss
		} else {
			copy := p
			grouped[key] = &copy
			totalAmt[key] = p.AvgPrice * float64(p.Quantity)
		}
	}

	result := make([]broker.Position, 0, len(grouped))
	for _, p := range grouped {
		result = append(result, *p)
	}
	return result
}
