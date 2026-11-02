package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func makeRouter(backendURL string) http.Handler {
	return newRouter(backendURL)
}

func TestHealthz(t *testing.T) {
	os.Setenv("CARD_SCOUT_URL", "http://localhost")
	defer os.Unsetenv("CARD_SCOUT_URL")

	router := makeRouter("http://localhost")
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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
	router := makeRouter("http://localhost")
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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

func TestUnknownRouteReturns404Not200(t *testing.T) {
	router := makeRouter("http://localhost")

	cases := []struct {
		method, path string
	}{
		{"GET", "/api/not-a-card-scout-route"},
		{"POST", "/api/secret"},
		{"GET", "/api/cards"},
		{"POST", "/api/batch/delete"},
		{"GET", "/proxy?url=https://evil.com"},
		{"GET", "/assets/item-001/extra"},
		{"GET", "/assets/item-abcd"},
		{"GET", "/assets/item-01"},
		{"POST", "/api/cards/item-1/research"},
		{"POST", "/api/cards/item-1234/research"},
		{"GET", "/random-page"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Result().StatusCode != 404 {
				t.Errorf("expected 404 for %s %s, got %d", tc.method, tc.path, w.Result().StatusCode)
			}
		})
	}
}

func TestProxyForwardsStatusAndBody(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`{"error":"teapot"}`))
	}))
	defer backend.Close()

	router := makeRouter(backend.URL)
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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

	router := makeRouter(backend.URL)
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestProxyPreservesRedirectNoFollow(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/other")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer backend.Close()

	router := makeRouter(backend.URL)
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", resp.StatusCode)
	}
}

func TestProxyImagePreservesContentType(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer backend.Close()

	router := makeRouter(backend.URL)
	req := httptest.NewRequest("GET", "/assets/item-001", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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

func TestConcurrentImageRequestsNoDataRace(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer backend.Close()

	router := makeRouter(backend.URL)

	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/assets/item-001", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Result().StatusCode != 200 {
				t.Errorf("request %d: expected 200, got %d", i, w.Result().StatusCode)
			}
		}(i)
	}
	wg.Wait()
}

func TestResearchRejectsInvalidPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("backend should not be called")
	}))
	defer backend.Close()

	router := makeRouter(backend.URL)

	cases := []string{
		"/api/cards/item-001/extra",
		"/api/cards/item-001/",
		"/api/cards/foo/research",
		"/api/cards/item-abc/research",
		"/api/cards/item-01/research",
		"/api/cards/item-12345/research",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

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

	router := makeRouter(backend.URL)
	req := httptest.NewRequest("POST", "/api/cards/item-042/research", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

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
	router := makeRouter("http://127.0.0.1:19999")
	req := httptest.NewRequest("GET", "/api/batch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 502 {
		t.Fatalf("expected 502 for unavailable upstream, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "unavailable") {
		t.Fatal("expected 'unavailable' in error")
	}
}

func TestContextTimeoutOnProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer backend.Close()

	// construct proxy directly for timeout test with a short client
	p := &proxyHandler{
		baseURL: backend.URL,
		normalClient: &http.Client{
			Timeout: 50 * time.Millisecond,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

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

	router := makeRouter(backend.URL)
	req := httptest.NewRequest("POST", "/api/cards/item-001/research", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Pok") {
		t.Fatalf("expected Pokémon error, got %q", string(body))
	}
}

func TestAssetPathExactThreeDigits(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer backend.Close()

	router := makeRouter(backend.URL)

	// Valid: exactly 3 digits
	req := httptest.NewRequest("GET", "/assets/item-039", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Result().StatusCode != 200 {
		t.Fatalf("expected 200 for /assets/item-039, got %d", w.Result().StatusCode)
	}

	// Invalid: not 3 digits
	for _, path := range []string{"/assets/item-01", "/assets/item-1234", "/assets/item-abcd"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Result().StatusCode != 404 {
			t.Errorf("expected 404 for %s, got %d", path, w.Result().StatusCode)
		}
	}
}