package kis

import (
	"context"

	internaladapter "github.com/smallfish06/krsec/internal/kis/adapter"
	"github.com/smallfish06/krsec/pkg/broker"
	tokencache "github.com/smallfish06/krsec/pkg/token"
)

// Adapter is the public KIS adapter contract.
type Adapter interface {
	broker.Broker
	GetOrder(ctx context.Context, orderID string) (*broker.OrderResult, error)
	GetOrderFills(ctx context.Context, orderID string) ([]broker.OrderFill, error)
	GetInstrument(ctx context.Context, market, symbol string) (*broker.Instrument, error)
	CallEndpoint(ctx context.Context, method, path, trID string, fields map[string]string) (map[string]interface{}, error)
	BootstrapSymbols(ctx context.Context) (int, error)
	ReloadSymbols(ctx context.Context) (int, error)
}

// Options configures KIS adapter internals.
type Options struct {
	TokenManager    tokencache.Manager
	OrderContextDir string
}

// NewAdapterWithOptions creates a KIS adapter with injectable options.
func NewAdapterWithOptions(sandbox bool, accountID string, opts Options) Adapter {
	return internaladapter.NewAdapterWithOptions(sandbox, accountID, internaladapter.Options{
		TokenManager:    opts.TokenManager,
		OrderContextDir: opts.OrderContextDir,
	})
}
