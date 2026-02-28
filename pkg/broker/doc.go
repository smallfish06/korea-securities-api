// Package broker defines the common interface and types shared across all
// broker adapters.
//
// The central [Broker] interface provides methods for authentication, quotes,
// OHLCV data, balance/position queries, and order management. Concrete
// implementations live under internal/kis (and future brokers).
//
// Types such as [Quote], [Balance], [Position], [Instrument], [OHLCV],
// [OrderRequest], and [OrderResult] are broker-agnostic and safe for external use.
package broker
