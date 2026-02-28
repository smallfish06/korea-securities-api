package kiwoom

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed catalog.json
var embeddedCatalogJSON []byte

// Catalog describes the Kiwoom OpenAPI catalog scraped from official docs.
type Catalog struct {
	Source      string    `json:"source"`
	GeneratedAt string    `json:"generatedAt"`
	Count       int       `json:"count"`
	APIs        []APISpec `json:"apis"`
}

// APISpec describes one Kiwoom API endpoint/tr-id.
type APISpec struct {
	APIID          string `json:"apiId"`
	Name           string `json:"name"`
	JobTpCode      string `json:"jobTpCode"`
	JobName        string `json:"jobName"`
	Method         string `json:"method"`
	RealDomain     string `json:"realDomain"`
	SimulateDomain string `json:"simulateDomain"`
	Path           string `json:"path"`
	Format         string `json:"format"`
	ContentType    string `json:"contentType"`
	Request        IOSpec `json:"request"`
	Response       IOSpec `json:"response"`
}

// IOSpec describes request/response schema rows shown in the guide.
type IOSpec struct {
	ExReqBody string      `json:"exReqBody"`
	InOutTp   string      `json:"inOutTp"`
	BodyData  []FieldSpec `json:"bodyData"`
	Header    []FieldSpec `json:"header"`
}

// FieldSpec describes one field metadata row in Kiwoom docs.
type FieldSpec struct {
	ItemID       string  `json:"itemId"`
	SortOrd      float64 `json:"sortOrd"`
	SampData     string  `json:"sampData"`
	InputOutput  string  `json:"inptOutputTp"`
	RequiredYN   string  `json:"esntYn"`
	Type         string  `json:"type"`
	HeadBodyType string  `json:"headBodyTp"`
	ItemName     string  `json:"itemNm"`
	Length       string  `json:"lngt"`
	Description  string  `json:"itemDc"`
}

var (
	catalogOnce sync.Once
	catalogErr  error
	catalogData Catalog
	catalogMap  map[string]APISpec
)

func loadCatalog() error {
	catalogOnce.Do(func() {
		if err := json.Unmarshal(embeddedCatalogJSON, &catalogData); err != nil {
			catalogErr = fmt.Errorf("parse embedded kiwoom catalog: %w", err)
			return
		}
		catalogMap = make(map[string]APISpec, len(catalogData.APIs))
		for _, spec := range catalogData.APIs {
			id := strings.TrimSpace(spec.APIID)
			if id == "" {
				continue
			}
			catalogMap[strings.ToLower(id)] = spec
		}
	})
	return catalogErr
}

// ListAPISpecs returns the full Kiwoom API catalog.
func ListAPISpecs() ([]APISpec, error) {
	if err := loadCatalog(); err != nil {
		return nil, err
	}
	out := make([]APISpec, len(catalogData.APIs))
	copy(out, catalogData.APIs)
	return out, nil
}

// LookupAPISpec finds one API spec by api-id (case-insensitive).
func LookupAPISpec(apiID string) (APISpec, bool, error) {
	if err := loadCatalog(); err != nil {
		return APISpec{}, false, err
	}
	spec, ok := catalogMap[strings.ToLower(strings.TrimSpace(apiID))]
	return spec, ok, nil
}
