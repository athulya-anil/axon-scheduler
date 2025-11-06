package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CacheClient handles communication with the semantic cache service
type CacheClient struct {
	baseURL    string
	httpClient *http.Client
}

// CacheRequest represents a request to add an entry to the cache
type CacheRequest struct {
	Query      string `json:"query"`
	Result     string `json:"result"`
	TTLSeconds int    `json:"ttl_seconds"`
}

// SearchRequest represents a cache search request
type SearchRequest struct {
	Query string `json:"query"`
}

// SearchResponse represents a cache search response
type SearchResponse struct {
	Hit        bool               `json:"hit"`
	Query      *string            `json:"query"`
	Result     *string            `json:"result"`
	Similarity *float64           `json:"similarity"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// StatsResponse represents cache statistics
type StatsResponse struct {
	TotalEntries   int     `json:"total_entries"`
	TotalQueries   int     `json:"total_queries"`
	CacheHits      int     `json:"cache_hits"`
	CacheMisses    int     `json:"cache_misses"`
	HitRate        float64 `json:"hit_rate"`
	HitRatePercent float64 `json:"hit_rate_percent"`
}

// NewCacheClient creates a new cache client
func NewCacheClient(baseURL string) *CacheClient {
	return &CacheClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Add adds an entry to the cache
func (c *CacheClient) Add(query, result string, ttlSeconds int) error {
	req := CacheRequest{
		Query:      query,
		Result:     result,
		TTLSeconds: ttlSeconds,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/cache",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to add to cache: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cache add failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Search searches the cache for similar queries
func (c *CacheClient) Search(query string) (*SearchResponse, error) {
	req := SearchRequest{
		Query: query,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/search",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search cache: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cache search failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}

// GetStats retrieves cache statistics
func (c *CacheClient) GetStats() (*StatsResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/stats")
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get stats failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

// Clear clears all cache entries
func (c *CacheClient) Clear() error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/cache", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cache clear failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Health checks if the cache service is healthy
func (c *CacheClient) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cache service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
