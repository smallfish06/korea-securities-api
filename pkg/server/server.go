package server

import (
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/smallfish06/korea-securities-api/internal/config"
	internalserver "github.com/smallfish06/korea-securities-api/internal/server"
	"github.com/smallfish06/korea-securities-api/pkg/broker"
)

// Account describes an externally supplied account/broker binding.
type Account struct {
	ID          string
	Name        string
	Broker      string
	Sandbox     bool
	Credentials broker.Credentials
}

// Options configures the public API server.
// External users can provide their own broker implementations through Brokers.
type Options struct {
	Host     string
	Port     int
	Accounts []Account
	Brokers  map[string]broker.Broker // account_id -> broker implementation
}

// Server wraps the internal HTTP server and exposes a stable public API.
type Server struct {
	inner *internalserver.Server
}

// New creates a server with externally supplied broker implementations.
func New(opts Options) *Server {
	inner := internalserver.NewWithBrokers(
		opts.Host,
		opts.Port,
		toInternalAccounts(opts.Accounts),
		opts.Brokers,
	)
	return &Server{inner: inner}
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	return s.inner.Run()
}

// App returns the underlying Fuego server for embedding or custom route composition.
func (s *Server) App() *fuego.Server {
	return s.inner.App()
}

func toInternalAccounts(accounts []Account) []config.AccountConfig {
	out := make([]config.AccountConfig, 0, len(accounts))
	for _, acc := range accounts {
		id := strings.TrimSpace(acc.ID)
		if id == "" {
			continue
		}
		out = append(out, config.AccountConfig{
			Name:      strings.TrimSpace(acc.Name),
			Broker:    strings.ToLower(strings.TrimSpace(acc.Broker)),
			Sandbox:   acc.Sandbox,
			AppKey:    strings.TrimSpace(acc.Credentials.AppKey),
			AppSecret: strings.TrimSpace(acc.Credentials.AppSecret),
			AccountID: id,
		})
	}
	return out
}
