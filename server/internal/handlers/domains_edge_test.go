package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// largeMockDomainsService returns a fixed list of 100 domain strings.
type largeMockDomainsService struct{}

func (l *largeMockDomainsService) KBDomains() []string {
	domains := make([]string, 100)
	for i := range domains {
		domains[i] = strings.ToLower(string(rune('a'+i%26))) + string(rune('0'+i/26)) + "_domain"
	}
	return domains
}

// TestDomainsEdgeLargeListReturns100Entries verifies that a service returning
// 100 domains is correctly forwarded to the client — no truncation or error.
func TestDomainsEdgeLargeListReturns100Entries(t *testing.T) {
	t.Parallel()
	app := newDomainsApp(&largeMockDomainsService{})

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var body map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	got, ok := body["domains"]
	if !ok {
		t.Fatal("missing 'domains' key in response")
	}
	if len(got) != 100 {
		t.Errorf("domains: want 100, got %d", len(got))
	}
}

// TestDomainsEdgeContentTypeIsAlwaysJSON verifies that the handler always
// sets Content-Type: application/json regardless of whether the list is empty
// or populated.
func TestDomainsEdgeContentTypeIsAlwaysJSON(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		service DomainsService
	}{
		{"empty list", &mockDomainsService{domains: []string{}}},
		{"nil list", &mockDomainsService{domains: nil}},
		{"populated list", &mockDomainsService{domains: []string{"hvac", "3d_printing"}}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := newDomainsApp(tc.service)

			req := httptest.NewRequest(http.MethodGet, "/domains", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("test: %v", err)
			}
			ct := resp.Header.Get("Content-Type")
			if !strings.Contains(ct, "application/json") {
				t.Errorf("[%s] Content-Type: want application/json, got %q", tc.name, ct)
			}
		})
	}
}

// TestDomainsEdgeNilServiceProducesEmptyArray verifies that when the handler
// is constructed with a nil DomainsService (rag == nil) the JSON response is
// {"domains":[]} and not {"domains":null}.
func TestDomainsEdgeNilServiceProducesEmptyArray(t *testing.T) {
	t.Parallel()
	// Pass nil DomainsService — DomainsHandler handles nil rag gracefully.
	app := newDomainsApp(nil)

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	domainsRaw, ok := raw["domains"]
	if !ok {
		t.Fatal("missing 'domains' key")
	}
	if string(domainsRaw) == "null" {
		t.Errorf("nil service: want [], got null")
	}
	var arr []string
	if err := json.Unmarshal(domainsRaw, &arr); err != nil {
		t.Errorf("domains is not a valid JSON array: %v", err)
	}
	if len(arr) != 0 {
		t.Errorf("nil service: want empty array, got %d elements", len(arr))
	}
}

// TestDomainsEdgeEmptySliceProducesArrayNotNull verifies that when
// KBDomains() returns an empty (non-nil) []string the JSON output is still
// [] and not null.
func TestDomainsEdgeEmptySliceProducesArrayNotNull(t *testing.T) {
	t.Parallel()
	app := newDomainsApp(&mockDomainsService{domains: []string{}})

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	domainsRaw := raw["domains"]
	if string(domainsRaw) == "null" {
		t.Errorf("empty slice: want [], got null")
	}
}

// TestDomainsEdgeResponseBodyIsValidJSON verifies that the response body is
// always decodable as valid JSON (regression guard for marshalling errors).
func TestDomainsEdgeResponseBodyIsValidJSON(t *testing.T) {
	t.Parallel()
	app := newDomainsApp(&mockDomainsService{domains: []string{"robotics", "plumbing"}})

	req := httptest.NewRequest(http.MethodGet, "/domains", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if _, ok := raw["domains"]; !ok {
		t.Error("expected top-level 'domains' key in JSON response")
	}
}
