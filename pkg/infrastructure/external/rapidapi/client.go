package rapidapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		timeout = 30 * time.Second
	}

	return &LinkedInClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchProfileByURN fetches a LinkedIn profile by URN from RapidAPI
func (c *LinkedInClient) FetchProfileByURN(
	ctx context.Context,
	urn string,
) (*LinkedInProfile, []byte, error) {
	// Construct the API URL - adjust based on actual RapidAPI endpoint
	url := fmt.Sprintf("%s/all-profile-data", c.baseURL)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("username", urn)
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

	// Check for error response format
	var errorResp APIErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && !errorResp.Success {
		// API returned an error in the expected format
		return nil, nil, fmt.Errorf("API error: %s", errorResp.Message)
	}

	// Parse response
	var profile LinkedInProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, body, fmt.Errorf("failed to parse response: %w", err)
	}

	return &profile, body, nil
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

	// Check for error response format
	var errorResp APIErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && !errorResp.Success {
		// API returned an error in the expected format
		return nil, nil, fmt.Errorf("API error: %s", errorResp.Message)
	}

	// Parse response
	var profile LinkedInProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, body, fmt.Errorf("failed to parse response: %w", err)
	}

	return &profile, body, nil
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

	// Check for error response format
	var errorResp APIErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && !errorResp.Success {
		// API returned an error in the expected format
		return nil, nil, fmt.Errorf("API error: %s", errorResp.Message)
	}

	var profile LinkedInProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, body, fmt.Errorf("failed to parse response: %w", err)
	}

	return &profile, body, nil
}
