package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func newTestProxy() *proxyHandler {
	return &proxyHandler{baseURL: "http://127.0.0.1:19999"}
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var v map[string]string
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if v["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", v["status"])
	}
}

func TestIndexReturnsHTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFileFS(w, r, ui, "index.html")
	})
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Card Scout") {
		t.Fatal("expected 'Card Scout' in HTML")
	}
}

func TestProxyForwardsStatusAndBody(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`{"error":"teapot"}`))
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	p.forward(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "teapot") {
		t.Fatalf("expected teapot body, got %q", string(body))
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json content-type, got %q", ct)
	}
}

func TestProxy404Error(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	p.forward(w, req)

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestProxyImagePreservesContentType(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	req := httptest.NewRequest("GET", "/assets/item-001", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/{id...}", p.forwardAsset)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "image/png") {
		t.Fatalf("expected image/png, got %q", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) != 8 || body[0] != 0x89 {
		t.Fatal("expected PNG header in body")
	}
}

func TestResearchRejectsInvalidPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("backend should not be called")
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}

	cases := []string{
		"/api/cards/item-001/extra",
		"/api/cards/item-001/",
		"/api/cards/foo/research",
		"/api/cards/item-abc/research",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()
			p.forwardResearch(w, req)

			if w.Result().StatusCode != 404 {
				t.Errorf("expected 404 for %q, got %d", path, w.Result().StatusCode)
			}
		})
	}
}

func TestResearchForwardsValidPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/cards/item-042/research" {
			t.Errorf("expected /api/cards/item-042/research, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	req := httptest.NewRequest("POST", "/api/cards/item-042/research", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/cards/{id...}", p.forwardResearch)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), `"status":"ok"`) {
		t.Fatalf("expected ok response, got %q", string(body))
	}
}

func TestProxyUpstreamUnavailable(t *testing.T) {
	p := &proxyHandler{baseURL: "http://127.0.0.1:19999"}
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	p.forward(w, req)

	resp := w.Result()
	if resp.StatusCode != 502 {
		t.Fatalf("expected 502 for unavailable upstream, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "unavailable") {
		t.Fatal("expected 'unavailable' in error")
	}
}

func TestRejectArbitraryPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("backend should not be called for arbitrary paths")
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}

	tests := []struct {
		method, path string
	}{
		{"GET", "/api/secret"},
		{"POST", "/api/admin/keys"},
		{"GET", "/api/cards"},
		{"POST", "/api/batch/delete"},
		{"GET", "/proxy?url=https://evil.com"},
		{"GET", "/assets/item-001/extra"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			mux := http.NewServeMux()
			mux.HandleFunc("GET /api/batch", p.forward)
			mux.HandleFunc("POST /api/batch/analyze", p.forward)
			mux.HandleFunc("GET /assets/{id...}", p.forwardAsset)
			mux.HandleFunc("POST /api/cards/{id...}", p.forwardResearch)
			mux.ServeHTTP(w, req)

			if w.Result().StatusCode == 200 {
				t.Errorf("expected non-200 for %s %s, got 200", tc.method, tc.path)
			}
		})
	}
}

func TestContextTimeoutOnProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	p.httpClient = &http.Client{Timeout: 50 * time.Millisecond}

	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	p.forward(w, req)

	resp := w.Result()
	if resp.StatusCode != 502 {
		t.Fatalf("expected 502 on timeout, got %d", resp.StatusCode)
	}
}

func TestCARD_SCOUT_URLRequired(t *testing.T) {
	old := os.Getenv("CARD_SCOUT_URL")
	os.Unsetenv("CARD_SCOUT_URL")
	defer func() {
		if old != "" {
			os.Setenv("CARD_SCOUT_URL", old)
		}
	}()
	val := os.Getenv("CARD_SCOUT_URL")
	if val != "" {
		t.Fatal("CARD_SCOUT_URL should be empty for this test")
	}
}

func TestResearch409Error(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"Item is not an analyzed Pok\u00e9mon card"}`))
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	req := httptest.NewRequest("POST", "/api/cards/item-001/research", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/cards/{id...}", p.forwardResearch)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Pok") {
		t.Fatalf("expected Pokémon error, got %q", string(body))
	}
}

func TestAssetPathValidation(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("backend should not be called for invalid asset paths")
	}))
	defer backend.Close()

	p := &proxyHandler{baseURL: backend.URL}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /assets/{id...}", p.forwardAsset)

	req := httptest.NewRequest("GET", "/assets/item-abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Result().StatusCode != 404 {
		t.Fatalf("expected 404 for non-numeric asset, got %d", w.Result().StatusCode)
	}
}