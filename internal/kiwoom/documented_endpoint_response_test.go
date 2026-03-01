package kiwoom

import (
	"testing"

	kiwoomspecs "github.com/smallfish06/krsec/pkg/kiwoom/specs"
)

func TestDocumentedEndpointResponseFactoryCoverage(t *testing.T) {
	t.Parallel()

	if got := kiwoomspecs.DocumentedEndpointResponseFactoryCount(); got != kiwoomspecs.DocumentedEndpointSpecCount() {
		t.Fatalf("factory count = %d, spec count = %d", got, kiwoomspecs.DocumentedEndpointSpecCount())
	}
}

func TestNewDocumentedEndpointResponse_KnownAndUnknown(t *testing.T) {
	t.Parallel()

	known := kiwoomspecs.NewDocumentedEndpointResponse(PathStockInfo, "ka10001")
	if known == nil {
		t.Fatal("known endpoint response factory returned nil")
	}

	unknown := kiwoomspecs.NewDocumentedEndpointResponse("/api/unknown/path", "zz99999")
	if unknown != nil {
		t.Fatalf("unknown endpoint response factory returned %#v", unknown)
	}
}
