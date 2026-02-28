package server

import (
	"testing"

	"github.com/smallfish06/korea-securities-api/internal/config"
)

func TestNew_WiresKiwoomBroker(t *testing.T) {
	cfg := &config.Config{
		Server:  config.ServerConfig{Host: "127.0.0.1", Port: 18081},
		Storage: config.StorageConfig{},
		Accounts: []config.AccountConfig{
			{
				Name:      "kiwoom-main",
				Broker:    "kiwoom",
				Sandbox:   true,
				AppKey:    "",
				AppSecret: "",
				AccountID: "1234567890",
			},
		},
	}

	s := New(cfg)
	brk := s.getBroker("1234567890")
	if brk == nil {
		t.Fatal("expected kiwoom broker to be wired")
	}
	if got := brk.Name(); got != "KIWOOM" {
		t.Fatalf("broker name = %q, want KIWOOM", got)
	}
}
