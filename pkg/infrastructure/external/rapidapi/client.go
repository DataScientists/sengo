package rapidapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sheng-go-backend/config"
	"strconv"
	"strings"
	"time"
)

// LinkedInProfile represents the full LinkedIn profile response from RapidAPI
type LinkedInProfile struct {
	URN           string                   `json:"urn"`
	Username      string                   `json:"username"`
	FirstName     string                   `json:"firstName"`
	LastName      string                   `json:"lastName"`
	Headline      string                   `json:"headline"`
	Geo           *GeoData                 `json:"geo"`
	Educations    []map[string]interface{} `json:"educations"`
	FullPositions []map[string]interface{} `json:"fullPositions"`
	Skills        []map[string]interface{} `json:"skills"`
}

// GeoData represents geographic information
type GeoData struct {
	Country     string `json:"country"`
	City        string `json:"city"`
	Full        string `json:"full"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"country_name"` // Keep for backward compatibility
	CityName    string `json:"city_name"`    // Keep for backward compatibility
}

// APIResponse represents the wrapper response from RapidAPI
type APIResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    *json.RawMessage `json:"data"`
}

// APIErrorResponse represents error responses from RapidAPI
type APIErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// RateLimitError represents a RapidAPI rate limit response (e.g., HTTP 429)
type RateLimitError struct {
	RetryAfter time.Duration
	StatusCode int
	Message    string
}

func (e *RateLimitError) Error() string {
	if e == nil {
		return ""
	}
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited (status=%d, retry after %v): %s", e.StatusCode, e.RetryAfter, e.Message)
	}
	return fmt.Sprintf("rate limited (status=%d): %s", e.StatusCode, e.Message)
}

// NotFoundError represents a 404 response from RapidAPI (profile not found)
type NotFoundError struct {
	URN     string
	Message string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("profile not found for URN %s: %s", e.URN, e.Message)
}

// LinkedInClient handles RapidAPI LinkedIn requests
type LinkedInClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewLinkedInClient creates a new RapidAPI LinkedIn client
func NewLinkedInClient() *LinkedInClient {
	cfg := config.C.RapidAPI

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second // Increased default timeout to 60 seconds
	}

	return &LinkedInClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// parseAPIResponse handles the different response formats from RapidAPI
// Returns (profile, rawBody, error)
func parseAPIResponse(body []byte) (*LinkedInProfile, []byte, error) {
	// First, try to parse as a wrapped response with success/message/data
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err == nil {
		// Check if it has the success field (indicates it's a wrapped response)
		var tempCheck map[string]interface{}
		json.Unmarshal(body, &tempCheck)

		if _, hasSuccess := tempCheck["success"]; hasSuccess {
			// This is a wrapped response format
			if !apiResp.Success {
				// Error response
				log.Printf("RapidAPI Error Response: success=false, message=%s", apiResp.Message)
				// Check if the message indicates an invalid/not-found profile
				msgLower := strings.ToLower(apiResp.Message)
				if strings.Contains(msgLower, "not valid linkedin profile") ||
					strings.Contains(msgLower, "can't be accessed") ||
					strings.Contains(msgLower, "profile not found") {
					return nil, body, &NotFoundError{
						Message: apiResp.Message,
					}
				}
				return nil, body, fmt.Errorf("API error: %s", apiResp.Message)
			}

			// Success response with data wrapper
			if apiResp.Data == nil {
				return nil, body, fmt.Errorf("API error: success=true but data is null")
			}

			// Parse the profile from the data field
			var profile LinkedInProfile
			if err := json.Unmarshal(*apiResp.Data, &profile); err != nil {
				return nil, body, fmt.Errorf("failed to parse profile from data field: %w", err)
			}

			log.Printf("RapidAPI Success Response: parsed profile with data wrapper, username=%s", profile.Username)
			return &profile, body, nil
		}
	}

	// If not a wrapped response, try to parse directly as profile data
	var profile LinkedInProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, body, fmt.Errorf("failed to parse response as profile: %w", err)
	}

	log.Printf("RapidAPI Success Response: parsed direct profile data, username=%s", profile.Username)
	return &profile, body, nil
}

// FetchProfileByURN fetches a LinkedIn profile by URN from RapidAPI
func (c *LinkedInClient) FetchProfileByURN(
	ctx context.Context,
	urn string,
) (*LinkedInProfile, []byte, error) {
	// Construct the API URL - adjust based on actual RapidAPI endpoint

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("username", urn)
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Set("X-RapidAPI-Key", c.apiKey)
	req.Header.Set("X-RapidAPI-Host", "real-time-people-company-data.p.rapidapi.com")

	// Log the request details
	log.Printf("RapidAPI Request: %s %s", req.Method, req.URL.String())
	log.Printf(
		"RapidAPI Headers: Host=%s, Key=%s...",
		req.Header.Get("X-RapidAPI-Host"),
		c.apiKey[:10],
	)

	// Execute request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("RapidAPI Error after %v: %v", time.Since(startTime), err)
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf(
		"RapidAPI Response received after %v: Status=%d",
		time.Since(startTime),
		resp.StatusCode,
	)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("RapidAPI Response body length: %d bytes", len(body))

	// Check status code
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		log.Printf("RapidAPI rate limit response: status=429 retryAfter=%v body=%s", retryAfter, string(body))
		return nil, body, &RateLimitError{
			RetryAfter: retryAfter,
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("RapidAPI profile not found: URN=%s body=%s", urn, string(body))
		return nil, body, &NotFoundError{
			URN:     urn,
			Message: string(body),
		}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("RapidAPI Error response: %s", string(body))
		return nil, body, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response using the multi-format handler
	return parseAPIResponse(body)
}

// FetchProfileByUsername fetches a LinkedIn profile by username from RapidAPI
func (c *LinkedInClient) FetchProfileByUsername(
	ctx context.Context,
	username string,
) (*LinkedInProfile, []byte, error) {
	// Construct the API URL - base URL with username query parameter
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("username", username)
	req.URL.RawQuery = q.Encode()

	// Add headers
	req.Header.Set("X-RapidAPI-Key", c.apiKey)
	req.Header.Set("X-RapidAPI-Host", "linkedin-api8.p.rapidapi.com")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, body, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response using the multi-format handler
	return parseAPIResponse(body)
}

// FetchProfileByURL fetches a LinkedIn profile by URL from RapidAPI (alternative method)
func (c *LinkedInClient) FetchProfileByURL(
	ctx context.Context,
	profileURL string,
) (*LinkedInProfile, []byte, error) {
	url := fmt.Sprintf("%s/get-profile-data-by-url", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("url", profileURL)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-RapidAPI-Key", c.apiKey)
	req.Header.Set("X-RapidAPI-Host", "linkedin-api8.p.rapidapi.com")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, body, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response using the multi-format handler
	return parseAPIResponse(body)
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0
	}

	return time.Duration(seconds) * time.Second
}
