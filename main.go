package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:embed index.html
var ui embed.FS

var validAssetPath = regexp.MustCompile(`^/assets/item-\d{3}$`)
var validResearchPath = regexp.MustCompile(`^/api/cards/item-\d{3}/research$`)

func newRouter(baseURL string) http.Handler {
	proxy := &proxyHandler{
		baseURL:        strings.TrimRight(baseURL, "/"),
		normalClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		researchClient: &http.Client{
			Timeout: 210 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFileFS(w, r, ui, "index.html")
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("GET /api/batch", proxy.forward)
	mux.HandleFunc("POST /api/batch/analyze", proxy.forward)
	mux.HandleFunc("GET /assets/{id...}", proxy.forwardAsset)
	mux.HandleFunc("POST /api/cards/{id...}", proxy.forwardResearch)

	return withUnknown404(mux)
}

func withUnknown404(next http.Handler) http.Handler {
	known := map[string]bool{
		"/": true, "/healthz": true,
		"/api/batch": true,
		"/api/batch/analyze": true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if known[path] {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(path, "/assets/") {
			if validAssetPath.MatchString(path) {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		if strings.HasPrefix(path, "/api/cards/") {
			if validResearchPath.MatchString(path) && r.Method == "POST" {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})
}

type proxyHandler struct {
	baseURL        string
	normalClient   *http.Client
	researchClient *http.Client
}

func (p *proxyHandler) forward(w http.ResponseWriter, r *http.Request) {
	targetPath := r.URL.Path
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	u, err := url.Parse(p.baseURL + targetPath)
	if err != nil {
		http.Error(w, `{"error":"invalid target URL"}`, http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, u.String(), r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to create proxy request"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", r.Header.Get("Accept"))

	resp, err := p.normalClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		if k != "Content-Length" {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (p *proxyHandler) forwardAsset(w http.ResponseWriter, r *http.Request) {
	if !validAssetPath.MatchString(r.URL.Path) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	p.forward(w, r)
}

func (p *proxyHandler) forwardResearch(w http.ResponseWriter, r *http.Request) {
	if !validResearchPath.MatchString(r.URL.Path) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	targetPath := r.URL.Path
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	u, err := url.Parse(p.baseURL + targetPath)
	if err != nil {
		http.Error(w, `{"error":"invalid target URL"}`, http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, u.String(), r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to create proxy request"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("Accept", r.Header.Get("Accept"))

	resp, err := p.researchClient.Do(req)
	if err != nil {
		http.Error(w, `{"error":"upstream unavailable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		if k != "Content-Length" {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	cardScoutURL := os.Getenv("CARD_SCOUT_URL")
	if cardScoutURL == "" {
		log.Fatal("CARD_SCOUT_URL is not set")
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      newRouter(cardScoutURL),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 240 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("Card Scout starting on %s", listenAddr)
	log.Fatal(srv.ListenAndServe())
}

func init() {
	http.DefaultTransport = &http.Transport{
		MaxIdleConns:        20,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}
}