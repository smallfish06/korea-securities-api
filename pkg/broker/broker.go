package broker

import "context"

// Broker is the common interface for all broker adapters
type Broker interface {
	// Name returns the broker name
	Name() string

	// Authenticate authenticates with the broker and returns a token
	Authenticate(ctx context.Context, creds Credentials) (*Token, error)

	// GetQuote retrieves a quote for a given market and symbol
	GetQuote(ctx context.Context, market, symbol string) (*Quote, error)

	// GetOHLCV retrieves OHLCV data for a given market and symbol
	GetOHLCV(ctx context.Context, market, symbol string, opts OHLCVOpts) ([]OHLCV, error)

	// GetBalance retrieves account balance
	GetBalance(ctx context.Context, accountID string) (*Balance, error)

	// GetPositions retrieves account positions
	GetPositions(ctx context.Context, accountID string) ([]Position, error)

	// PlaceOrder places a new order
	PlaceOrder(ctx context.Context, req OrderRequest) (*OrderResult, error)

	// CancelOrder cancels an order
	CancelOrder(ctx context.Context, orderID string) error

	// ModifyOrder modifies an existing order
	ModifyOrder(ctx context.Context, orderID string, req ModifyOrderRequest) (*OrderResult, error)
}
