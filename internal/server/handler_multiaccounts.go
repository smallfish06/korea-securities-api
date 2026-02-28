package server

import (
	"context"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/smallfish06/korea-securities-api/pkg/broker"
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

	for _, account := range s.accounts {
		brk := s.getBroker(account.AccountID)
		if brk == nil {
			continue
		}

		balance, err := brk.GetBalance(ctx, account.AccountID)
		if err != nil {
			// 에러가 발생해도 계속 진행
			continue
		}

		balances = append(balances, *balance)
		totalAssets += balance.TotalAssets
		totalCash += balance.Cash
		totalProfitLoss += balance.ProfitLoss
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

// Authenticate all accounts
func (s *Server) authenticateAll(ctx context.Context) error {
	for _, account := range s.accounts {
		brk := s.getBroker(account.AccountID)
		if brk == nil {
			continue
		}

		creds := broker.Credentials{
			AppKey:    account.AppKey,
			AppSecret: account.AppSecret,
		}

		if _, err := brk.Authenticate(ctx, creds); err != nil {
			return err
		}
	}
	return nil
}
