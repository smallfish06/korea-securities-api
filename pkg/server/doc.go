// Package server provides a ready-to-use HTTP API server that can be wired
// with any [github.com/smallfish06/krsec/pkg/broker.Broker] implementation.
//
// Use [New] to create a server with externally supplied broker instances,
// then call [Server.Run] to start listening. The server exposes REST endpoints
// for quotes, orders, accounts, instruments, and an auto-generated OpenAPI spec.
package server
