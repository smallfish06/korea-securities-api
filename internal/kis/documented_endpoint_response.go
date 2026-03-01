package kis

import (
	"strings"

	kisspecs "github.com/smallfish06/krsec/pkg/kis/specs"
)

// DocumentedEndpointResponse is the typed response contract for documented KIS endpoints.
type DocumentedEndpointResponse = kisspecs.DocumentedEndpointResponse

// DocumentedResponseBase holds common KIS response status fields.
type DocumentedResponseBase = kisspecs.DocumentedResponseBase

// DocumentedSlice accepts both object and array payloads and normalizes them to a typed slice.
type DocumentedSlice[T any] = kisspecs.DocumentedSlice[T]

// NewDocumentedEndpointResponse returns a typed response object for the endpoint path.
func NewDocumentedEndpointResponse(path string) DocumentedEndpointResponse {
	return kisspecs.NewDocumentedEndpointResponse(strings.TrimSpace(path))
}

// DocumentedEndpointResponseFactoryCount returns the number of typed documented endpoint responses.
func DocumentedEndpointResponseFactoryCount() int {
	return kisspecs.DocumentedEndpointResponseFactoryCount()
}
