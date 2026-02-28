package server

import "github.com/smallfish06/kr-broker-api/pkg/broker"

// BrokerWithExtras is a test-oriented interface for server handlers.
//
//go:generate go run github.com/vektra/mockery/v3@v3.6.4 --config ../../.mockery.yml
//nolint:revive // this interface intentionally composes optional handler capabilities.
type BrokerWithExtras interface {
	broker.Broker
	orderGetter
	orderFillsGetter
	instrumentGetter
}
