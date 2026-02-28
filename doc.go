// Package krbrokapi provides a unified gateway for Korean securities broker APIs.
//
// kr-broker-api abstracts away the differences between Korean broker REST APIs
// (KIS, Kiwoom, LS, etc.) behind a single, consistent interface. It can be used
// as a standalone HTTP server or embedded as a Go library.
//
// # Standalone server
//
//	make build && ./bin/kr-broker -config config.yaml
//
// # Library usage
//
// Import [github.com/smallfish06/kr-broker-api/pkg/broker] for the common
// interface and types, and [github.com/smallfish06/kr-broker-api/pkg/server]
// to spin up an HTTP server with your own broker implementations.
package krbrokapi
