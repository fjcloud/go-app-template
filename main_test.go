package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cardScoutURL = "http://localhost:99999"
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/api/batch", handleBatch)
	mux.HandleFunc("/api/batch/analyze", handleAnalyze)
	mux.HandleFunc("/assets/", handleAssets)
	mux.HandleFunc("/api/cards/", handleCardResearch)
	mux.HandleFunc("/", handleUI)
	return httptest.NewServer(mux)
}

func assertStatus(t *testing.T, code int, expected int, context string) {
	t.Helper()
	if code != expected {
		t.Errorf("%s: expected %d, got %d", context, expected, code)
	}
}

func assertJSONError(t *testing.T, resp *http.Response) {
	t.Helper()
	body, _ := io.ReadAll(resp.Body)
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content-type, got %s", ct)
	}
	if !strings.Contains(string(body), `"error"`) {
		t.Errorf("expected JSON error body, got: %s", body)
	}
}

func TestHealthz(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp.StatusCode, http.StatusOK, "healthz GET")
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"status":"ok"}` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestHealthzMethodNotAllowed(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req, _ := http.NewRequest(method, ts.URL+"/healthz", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed, "healthz "+method)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestBatchGetMethod(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req, _ := http.NewRequest(method, ts.URL+"/api/batch", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed, "/api/batch "+method)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestBatchGetProxyUpstreamError(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/batch")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp.StatusCode, http.StatusBadGateway, "/api/batch GET proxy")
	assertJSONError(t, resp)
}

func TestAnalyzeMethodValidation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req, _ := http.NewRequest(method, ts.URL+"/api/batch/analyze", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed, "/api/batch/analyze "+method)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestUI(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp.StatusCode, http.StatusOK, "GET /")
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html, got %s", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<!DOCTYPE html>") {
		t.Error("response does not contain HTML doctype")
	}
	if !strings.Contains(string(body), "Card Scout") {
		t.Error("response does not contain Card Scout title")
	}
}

func TestUINotFoundArbitraryPath(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp.StatusCode, http.StatusNotFound, "arbitrary path")
}

func TestUIMethodNotAllowed(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req, _ := http.NewRequest(method, ts.URL+"/", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed, "/ "+method)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestAssetExactItemNumeric(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/assets/item-001")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp.StatusCode, http.StatusBadGateway, "valid asset path")
	assertJSONError(t, resp)
}

func TestAssetPathInvalidPrefixes(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	paths := []string{
		"/assets/other-file",
		"/assets/item-abc",
		"/assets/item-",
		"/assets/item-001a",
		"/assets/item-1",
		"/assets/item-0001",
		"/assets/image.jpg",
	}
	for _, p := range paths {
		resp, err := http.Get(ts.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		assertStatus(t, resp.StatusCode, http.StatusNotFound, "invalid asset prefix: "+p)
		assertJSONError(t, resp)
	}
}

func TestAssetPathTraversalSanitized(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	paths := []string{
		"/assets/../../etc/passwd",
		"/assets/item-001/../secret",
	}
	for _, p := range paths {
		resp, err := http.Get(ts.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 for traversal path %s, got %d", p, resp.StatusCode)
		}
	}
}

func TestAssetMethodNotAllowed(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req, _ := http.NewRequest(method, ts.URL+"/assets/item-001", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed, "assets "+method)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestCardResearchExactPattern(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	var forwarded atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()
	cardScoutURL = upstream.URL
	// Observe the mux's canonical-path redirect without following it as a GET.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	invalidPaths := []string{
		"/api/cards/item-001",
		"/api/cards/item-001/research/extra",
		"/api/cards/item-abc/research",
		"/api/cards/item-/research",
		"/api/cards/item-001a/research",
		"/api/cards/item-1/research",
		"/api/cards/item-0001/research",
		"/api/cards/bad-001/research",
		"/api/cards/item-001/research%00",
		"/api/cards/item-001/../research",
	}
	for _, p := range invalidPaths {
		req, _ := http.NewRequest(http.MethodPost, ts.URL+p, nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(p, "/../") {
			assertStatus(t, resp.StatusCode, http.StatusMovedPermanently, "canonical research path: "+p)
		} else {
			assertStatus(t, resp.StatusCode, http.StatusNotFound, "invalid research: "+p)
			assertJSONError(t, resp)
		}
		resp.Body.Close()
	}
	if forwarded.Load() != 0 {
		t.Fatalf("invalid research paths reached the backend %d times", forwarded.Load())
	}
}

func TestCardResearchMethodNotAllowed(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req, _ := http.NewRequest(method, ts.URL+"/api/cards/item-001/research", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed, "research "+method)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestArbitraryProxyRejection(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	paths := []string{
		"/api/unknown",
		"/api/cards/../../secret",
		"/admin",
		"/.env",
		"/wp-admin",
	}
	for _, p := range paths {
		resp, err := http.Get(ts.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusNotFound, "arbitrary: "+p)
		resp.Body.Close()
	}
}

func TestProxyErrorJSONBody(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	paths := []struct {
		url    string
		method string
	}{
		{"/api/batch", http.MethodGet},
		{"/api/batch/analyze", http.MethodPost},
		{"/api/cards/item-001/research", http.MethodPost},
		{"/assets/item-001", http.MethodGet},
	}
	for _, p := range paths {
		req, _ := http.NewRequest(p.method, ts.URL+p.url, nil)
		if p.method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		assertStatus(t, resp.StatusCode, http.StatusBadGateway, p.method+" "+p.url)
		assertJSONError(t, resp)
		resp.Body.Close()
	}
}

func TestHeadUI(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodHead, ts.URL+"/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp.StatusCode, http.StatusOK, "HEAD /")
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html, got %s", ct)
	}
}
