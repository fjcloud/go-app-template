package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed static/*
var static embed.FS

var (
	cardScoutURL string
	httpClient   = &http.Client{Timeout: 30 * time.Second}
)

func main() {
	cardScoutURL = os.Getenv("CARD_SCOUT_URL")
	if cardScoutURL == "" {
		log.Fatal("CARD_SCOUT_URL not set")
	}
	u, err := url.Parse(cardScoutURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		log.Fatalf("invalid CARD_SCOUT_URL: %v", cardScoutURL)
	}
	cardScoutURL = strings.TrimRight(cardScoutURL, "/")

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/api/batch", handleBatch)
	mux.HandleFunc("/api/batch/analyze", handleAnalyze)
	mux.HandleFunc("/assets/", handleAssets)
	mux.HandleFunc("/api/cards/", handleCardResearch)
	mux.HandleFunc("/", handleUI)

	log.Printf("signal-desk listening on %s, proxying to %s", listenAddr, cardScoutURL)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatal(err)
	}
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, cardScoutURL+"/api/batch")
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	proxyRequest(w, r, cardScoutURL+"/api/batch/analyze")
}

func handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path
	if !strings.HasPrefix(path, "/assets/item-") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	proxyRequest(w, r, cardScoutURL+path)
}

func handleCardResearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/cards/item-") || !strings.HasSuffix(path, "/research") {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	proxyRequest(w, r, cardScoutURL+path)
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	content, err := static.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "UI not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func proxyRequest(w http.ResponseWriter, r *http.Request, target string) {
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		http.Error(w, "proxy error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("proxy error to %s: %v", target, err)
		http.Error(w, `{"error":"upstream unreachable"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

