package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/smallfish06/korea-securities-api/internal/config"
	"github.com/smallfish06/korea-securities-api/internal/kis"
	kisadapter "github.com/smallfish06/korea-securities-api/internal/kis/adapter"
	"github.com/smallfish06/korea-securities-api/internal/kiwoom"
	kiwoomadapter "github.com/smallfish06/korea-securities-api/internal/kiwoom/adapter"
	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// Server represents the HTTP server
type Server struct {
	config   *config.Config
	router   *fuego.Server
	brokers  map[string]broker.Broker // account_id -> broker adapter
	accounts []config.AccountConfig
}

func newBaseServer(cfg *config.Config) *Server {
	host := strings.TrimSpace(cfg.Server.Host)
	if host == "" {
		host = "0.0.0.0"
	}
	port := cfg.Server.Port
	if port <= 0 {
		port = 8080
	}
	addr := fmt.Sprintf("%s:%d", host, port)

	r := fuego.NewServer(
		fuego.WithAddr(addr),
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				PrettyFormatJSON: true,
				Info: &openapi3.Info{
					Title:       "KR Broker API",
					Description: "Unified broker API over multiple securities broker adapters",
					Version:     "1.0.0",
				},
			}),
		),
	)

	s := &Server{
		config:   cfg,
		router:   r,
		brokers:  make(map[string]broker.Broker),
		accounts: cfg.Accounts,
	}

	s.routes()
	return s
}

// New creates a new server instance.
// This constructor wires built-in brokers from config (currently KIS, Kiwoom).
func New(cfg *config.Config) *Server {
	s := newBaseServer(cfg)

	kisTokenManager := kis.NewFileTokenManagerWithDir(cfg.Storage.TokenDir)
	kiwoomTokenManager := kiwoom.NewFileTokenManagerWithDir(cfg.Storage.TokenDir)

	// Initialize brokers for each account
	for _, account := range cfg.Accounts {
		var brk broker.Broker
		switch account.Broker {
		case "kis":
			adapter := kisadapter.NewAdapterWithOptions(account.Sandbox, account.AccountID, kisadapter.Options{
				TokenManager:    kisTokenManager,
				OrderContextDir: cfg.Storage.OrderContextDir,
			})
			creds := broker.Credentials{
				AppKey:    account.AppKey,
				AppSecret: account.AppSecret,
			}
			// Authenticate in background (don't block server start)
			go func(name string, a *kisadapter.Adapter, c broker.Credentials) {
				if _, err := a.Authenticate(context.Background(), c); err != nil {
					log.Printf("Warning: failed to authenticate account %s: %v", name, err)
				} else {
					log.Printf("Authenticated account %s", name)
				}
			}(account.Name, adapter, creds)

			// Bootstrap symbol master files in background.
			go func(name string, a *kisadapter.Adapter) {
				ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
				defer cancel()
				count, err := a.BootstrapSymbols(ctx)
				if err != nil {
					log.Printf("Warning: symbol bootstrap failed for account %s: %v", name, err)
				} else {
					log.Printf("Bootstrapped %d symbol records for account %s", count, name)
				}

				// Keep symbol master cache fresh (KIS master files change over time).
				ticker := time.NewTicker(24 * time.Hour)
				defer ticker.Stop()
				for range ticker.C {
					reloadCtx, reloadCancel := context.WithTimeout(context.Background(), 90*time.Second)
					count, err := a.ReloadSymbols(reloadCtx)
					reloadCancel()
					if err != nil {
						log.Printf("Warning: symbol reload failed for account %s: %v", name, err)
						continue
					}
					log.Printf("Reloaded %d symbol records for account %s", count, name)
				}
			}(account.Name, adapter)
			brk = adapter
		case "kiwoom":
			adapter := kiwoomadapter.NewAdapterWithOptions(account.Sandbox, account.AccountID, kiwoomadapter.Options{
				TokenManager:    kiwoomTokenManager,
				OrderContextDir: cfg.Storage.OrderContextDir,
			})
			creds := broker.Credentials{
				AppKey:    account.AppKey,
				AppSecret: account.AppSecret,
			}
			go func(name string, a *kiwoomadapter.Adapter, c broker.Credentials) {
				if _, err := a.Authenticate(context.Background(), c); err != nil {
					log.Printf("Warning: failed to authenticate account %s: %v", name, err)
				} else {
					log.Printf("Authenticated account %s", name)
				}
			}(account.Name, adapter, creds)
			brk = adapter
		default:
			log.Printf("Warning: unknown broker type: %s", account.Broker)
			continue
		}
		s.brokers[account.AccountID] = brk
	}

	return s
}

// NewWithBrokers creates a server with externally provided brokers.
// This constructor is used by the public pkg/server package for OSS extensibility.
func NewWithBrokers(host string, port int, accounts []config.AccountConfig, brokers map[string]broker.Broker) *Server {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "0.0.0.0"
	}
	if port <= 0 {
		port = 8080
	}

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: host,
			Port: port,
		},
		Accounts: accounts,
	}
	s := newBaseServer(cfg)
	for accountID, brk := range brokers {
		if brk == nil {
			continue
		}
		s.brokers[accountID] = brk
	}
	return s
}

// routes sets up HTTP routes
func (s *Server) routes() {
	fuego.Get(s.router, "/health", s.handleHealth, fuego.OptionTags("System"), fuego.OptionSummary("Health check"))

	// Auth
	fuego.Post(s.router, "/auth/token", s.handleAuthToken, fuego.OptionTags("Auth"), fuego.OptionSummary("Issue broker auth token"))

	// Quotes (legacy endpoint - uses first account's broker)
	fuego.Get(s.router, "/quotes/{market}/{symbol}", s.handleGetQuote, fuego.OptionTags("Quotes"), fuego.OptionSummary("Get latest quote"))
	fuego.Get(s.router, "/quotes/{market}/{symbol}/ohlcv", s.handleGetOHLCV, fuego.OptionTags("Quotes"), fuego.OptionSummary("Get OHLCV candles"))
	fuego.Get(s.router, "/instruments/{market}/{symbol}", s.handleGetInstrument, fuego.OptionTags("Instruments"), fuego.OptionSummary("Get instrument metadata"))

	// Multi-account endpoints
	fuego.Get(s.router, "/accounts", s.handleListAccounts, fuego.OptionTags("Accounts"), fuego.OptionSummary("List configured accounts"))
	fuego.Get(s.router, "/accounts/summary", s.handleAccountsSummary, fuego.OptionTags("Accounts"), fuego.OptionSummary("Get combined account summary"))
	fuego.Get(s.router, "/accounts/{account_id}/balance", s.handleGetBalance, fuego.OptionTags("Accounts"), fuego.OptionSummary("Get account balance"))
	fuego.Get(s.router, "/accounts/{account_id}/positions", s.handleGetPositions, fuego.OptionTags("Accounts"), fuego.OptionSummary("Get account positions"))

	// Orders
	fuego.Get(s.router, "/orders/{order_id}/fills", s.handleGetOrderFills, fuego.OptionTags("Orders"), fuego.OptionSummary("Get order fills"))
	fuego.Get(s.router, "/orders/{order_id}", s.handleGetOrder, fuego.OptionTags("Orders"), fuego.OptionSummary("Get order status"))
	fuego.Post(s.router, "/orders", s.handlePlaceOrder, fuego.OptionTags("Orders"), fuego.OptionSummary("Place order"))
	fuego.Delete(s.router, "/orders/{order_id}", s.handleCancelOrder, fuego.OptionTags("Orders"), fuego.OptionSummary("Cancel order"))
	fuego.Put(s.router, "/orders/{order_id}", s.handleModifyOrder, fuego.OptionTags("Orders"), fuego.OptionSummary("Modify order"))
}

// Run starts the HTTP server
func (s *Server) Run() error {
	log.Printf("Server listening on %s", s.router.Addr)
	return s.router.Run()
}

// App returns the underlying Fuego server for embedding/customization.
func (s *Server) App() *fuego.Server {
	return s.router
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c fuego.ContextNoBody) (map[string]interface{}, error) {
	c.SetStatus(http.StatusOK)
	return map[string]interface{}{
		"status":   "ok",
		"accounts": len(s.accounts),
	}, nil
}

// getBroker returns the broker for the given account ID
func (s *Server) getBroker(accountID string) broker.Broker {
	if brk, ok := s.brokers[accountID]; ok {
		return brk
	}
	// Try matching with/without product code suffix (e.g., "73027400" matches "73027400-01")
	for key, brk := range s.brokers {
		if strings.HasPrefix(key, accountID+"-") || strings.HasPrefix(accountID, key+"-") || strings.TrimSuffix(key, "-01") == strings.TrimSuffix(accountID, "-01") {
			return brk
		}
	}
	// If not found, return first broker (legacy compatibility)
	if len(s.brokers) > 0 {
		for _, brk := range s.brokers {
			return brk
		}
	}
	return nil
}

// getFirstBroker returns the first available broker (for legacy endpoints)
func (s *Server) getFirstBroker() broker.Broker {
	if len(s.accounts) > 0 {
		return s.getBroker(s.accounts[0].AccountID)
	}
	return nil
}

// Response represents a standard API response
type Response struct {
	OK     bool        `json:"ok"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
	Broker string      `json:"broker,omitempty"`
}

func respond(c interface{ SetStatus(int) }, status int, data Response) (Response, error) {
	c.SetStatus(status)
	return data, nil
}
