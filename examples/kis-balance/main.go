package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/smallfish06/kr-broker-api/internal/config"
	"github.com/smallfish06/kr-broker-api/internal/kis"
	kisadapter "github.com/smallfish06/kr-broker-api/internal/kis/adapter"
	"github.com/smallfish06/kr-broker-api/pkg/broker"
)

type result struct {
	AccountID   string            `json:"account_id"`
	AccountName string            `json:"account_name"`
	Broker      string            `json:"broker"`
	Sandbox     bool              `json:"sandbox"`
	Balance     *broker.Balance   `json:"balance,omitempty"`
	Positions   []broker.Position `json:"positions,omitempty"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	accountSelector := flag.String("account", "", "Account ID or account name (optional, default first account)")
	withPositions := flag.Bool("positions", true, "Include positions in output")
	timeout := flag.Duration("timeout", 20*time.Second, "Request timeout")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	account, err := selectAccount(cfg, *accountSelector)
	if err != nil {
		log.Fatalf("select account: %v", err)
	}

	tokenManager := kis.NewFileTokenManagerWithDir(cfg.Storage.TokenDir)
	adapter := kisadapter.NewAdapterWithOptions(account.Sandbox, account.AccountID, kisadapter.Options{
		TokenManager:    tokenManager,
		OrderContextDir: cfg.Storage.OrderContextDir,
	})
	creds := broker.Credentials{
		AppKey:    account.AppKey,
		AppSecret: account.AppSecret,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if _, err := adapter.Authenticate(ctx, creds); err != nil {
		log.Fatalf("authenticate: %v", err)
	}

	bal, err := adapter.GetBalance(ctx, account.AccountID)
	if err != nil {
		log.Fatalf("get balance: %v", err)
	}

	out := result{
		AccountID:   account.AccountID,
		AccountName: account.Name,
		Broker:      account.Broker,
		Sandbox:     account.Sandbox,
		Balance:     bal,
	}

	if *withPositions {
		pos, err := adapter.GetPositions(ctx, account.AccountID)
		if err != nil {
			log.Fatalf("get positions: %v", err)
		}
		out.Positions = pos
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatalf("marshal output: %v", err)
	}
	fmt.Println(string(data))
}

func selectAccount(cfg *config.Config, selector string) (config.AccountConfig, error) {
	if len(cfg.Accounts) == 0 {
		return config.AccountConfig{}, fmt.Errorf("no accounts configured")
	}
	if selector == "" {
		return cfg.Accounts[0], nil
	}
	for _, acc := range cfg.Accounts {
		if acc.AccountID == selector || acc.Name == selector {
			return acc, nil
		}
	}
	return config.AccountConfig{}, fmt.Errorf("account not found: %s", selector)
}
