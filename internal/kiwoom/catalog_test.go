package kiwoom

import "testing"

func TestCatalogHasExpectedCoverage(t *testing.T) {
	apis, err := ListAPISpecs()
	if err != nil {
		t.Fatalf("ListAPISpecs error: %v", err)
	}
	if len(apis) < 200 {
		t.Fatalf("api spec count = %d, want >= 200", len(apis))
	}

	required := []string{"au10001", "ka10001", "ka10081", "kt00018", "kt10000", "kt10003"}
	for _, id := range required {
		if _, ok, err := LookupAPISpec(id); err != nil {
			t.Fatalf("LookupAPISpec(%s) error: %v", id, err)
		} else if !ok {
			t.Fatalf("required api-id missing: %s", id)
		}
	}
}
