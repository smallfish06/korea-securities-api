package server

import (
	"net/http"

	"github.com/go-fuego/fuego"

	"github.com/smallfish06/krsec/pkg/broker"
)

// handleListAccounts handles GET /accounts
func (s *Server) handleListAccounts(c fuego.ContextNoBody) (Response, error) {
	accounts := make([]broker.AccountInfo, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, broker.AccountInfo{
			ID:     account.AccountID,
			Name:   account.Name,
			Broker: account.Broker,
		})
	}

	return respond(c, http.StatusOK, Response{
		OK:   true,
		Data: accounts,
	})
}

// handleAccountsSummary handles GET /accounts/summary
func (s *Server) handleAccountsSummary(c fuego.ContextNoBody) (Response, error) {
	ctx := c.Context()

	var totalAssets, totalCash, totalProfitLoss float64
	balances := make([]broker.Balance, 0, len(s.accounts))
	failed := 0

	for _, account := range s.accounts {
		brk, status, _ := s.resolveBrokerByAccountID(account.AccountID)
		if status != 0 || brk == nil {
			failed++
			continue
		}

		balance, err := brk.GetBalance(ctx, account.AccountID)
		if err != nil {
			// 에러가 발생해도 계속 진행
			failed++
			continue
		}

		balances = append(balances, *balance)
		totalAssets += balance.TotalAssets
		totalCash += balance.Cash
		totalProfitLoss += balance.ProfitLoss
	}

	if len(s.accounts) > 0 && len(balances) == 0 && failed > 0 {
		return respond(c, http.StatusServiceUnavailable, Response{
			OK:    false,
			Error: "failed to retrieve balances from all accounts",
		})
	}

	summary := broker.AccountSummary{
		TotalAssets:     totalAssets,
		TotalCash:       totalCash,
		TotalProfitLoss: totalProfitLoss,
		Accounts:        balances,
	}

	return respond(c, http.StatusOK, Response{
		OK:   true,
		Data: summary,
	})
}
