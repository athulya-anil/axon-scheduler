package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCacheClient_Health(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
		}
	}))
	defer server.Close()

	client := NewCacheClient(server.URL)
	err := client.Health()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestCacheClient_Add(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cache" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"entry_id":0,"message":"Entry added"}`))
		}
	}))
	defer server.Close()

	client := NewCacheClient(server.URL)
	err := client.Add("test query", "test result", 3600)
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}
}

func TestCacheClient_Search_Hit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"hit": true,
				"query": "test query",
				"result": "test result",
				"similarity": 0.95,
				"metadata": {}
			}`))
		}
	}))
	defer server.Close()

	client := NewCacheClient(server.URL)
	resp, err := client.Search("test query")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if !resp.Hit {
		t.Error("Expected cache hit, got miss")
	}

	if resp.Similarity == nil || *resp.Similarity != 0.95 {
		t.Errorf("Expected similarity 0.95, got %v", resp.Similarity)
	}
}

func TestCacheClient_Search_Miss(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"hit": false}`))
		}
	}))
	defer server.Close()

	client := NewCacheClient(server.URL)
	resp, err := client.Search("unknown query")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if resp.Hit {
		t.Error("Expected cache miss, got hit")
	}
}

func TestCacheClient_GetStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stats" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"total_entries": 10,
				"total_queries": 20,
				"cache_hits": 15,
				"cache_misses": 5,
				"hit_rate": 0.75,
				"hit_rate_percent": 75.0
			}`))
		}
	}))
	defer server.Close()

	client := NewCacheClient(server.URL)
	stats, err := client.GetStats()
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}

	if stats.TotalEntries != 10 {
		t.Errorf("Expected 10 entries, got %d", stats.TotalEntries)
	}

	if stats.HitRate != 0.75 {
		t.Errorf("Expected hit rate 0.75, got %f", stats.HitRate)
	}
}

func TestCacheClient_Clear(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cache" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"Cache cleared","entries_removed":5}`))
		}
	}))
	defer server.Close()

	client := NewCacheClient(server.URL)
	err := client.Clear()
	if err != nil {
		t.Errorf("Clear failed: %v", err)
	}
}
