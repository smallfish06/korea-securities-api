package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/smallfish06/korea-securities-api/internal/config"
	"github.com/smallfish06/korea-securities-api/internal/kiwoom"
	kiwoomadapter "github.com/smallfish06/korea-securities-api/internal/kiwoom/adapter"
	"github.com/smallfish06/korea-securities-api/pkg/broker"
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
	accountSelector := flag.String("account", "", "Kiwoom account ID or account name (optional, default first Kiwoom account)")
	withPositions := flag.Bool("positions", true, "Include positions in output")
	timeout := flag.Duration("timeout", 20*time.Second, "Request timeout")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	account, err := selectKiwoomAccount(cfg, *accountSelector)
	if err != nil {
		log.Fatalf("select account: %v", err)
	}

	tokenManager := kiwoom.NewFileTokenManagerWithDir(cfg.Storage.TokenDir)
	adapter := kiwoomadapter.NewAdapterWithOptions(account.Sandbox, account.AccountID, kiwoomadapter.Options{
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

func selectKiwoomAccount(cfg *config.Config, selector string) (config.AccountConfig, error) {
	if len(cfg.Accounts) == 0 {
		return config.AccountConfig{}, fmt.Errorf("no accounts configured")
	}

	selector = strings.TrimSpace(selector)
	if selector != "" {
		for _, acc := range cfg.Accounts {
			if strings.ToLower(strings.TrimSpace(acc.Broker)) != "kiwoom" {
				continue
			}
			if acc.AccountID == selector || acc.Name == selector {
				return acc, nil
			}
		}
		return config.AccountConfig{}, fmt.Errorf("kiwoom account not found: %s", selector)
	}

	for _, acc := range cfg.Accounts {
		if strings.ToLower(strings.TrimSpace(acc.Broker)) == "kiwoom" {
			return acc, nil
		}
	}
	return config.AccountConfig{}, fmt.Errorf("no kiwoom accounts configured")
}
