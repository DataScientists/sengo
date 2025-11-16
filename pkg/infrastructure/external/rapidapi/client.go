package rapidapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sheng-go-backend/config"
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
	CountryName string `json:"country_name"`
	CityName    string `json:"city_name"`
}

// APIResponse represents the wrapper response from RapidAPI
type APIResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    *json.RawMessage        `json:"data"`
}

// APIErrorResponse represents error responses from RapidAPI
type APIErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
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
