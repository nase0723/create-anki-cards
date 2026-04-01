package image

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type PexelsClient struct {
	apiKey string
}

func NewPexelsClient(apiKey string) *PexelsClient {
	return &PexelsClient{apiKey: apiKey}
}

type pexelsResponse struct {
	Photos []struct {
		Src struct {
			Medium string `json:"medium"`
		} `json:"src"`
	} `json:"photos"`
}

func (c *PexelsClient) Search(keywords string, count int) ([]string, error) {
	u := fmt.Sprintf("https://api.pexels.com/v1/search?query=%s&per_page=%d",
		url.QueryEscape(keywords), count)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Pexels API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Pexels API error: %s", string(body))
	}

	var result pexelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	urls := make([]string, 0, len(result.Photos))
	for _, photo := range result.Photos {
		urls = append(urls, photo.Src.Medium)
	}

	return urls, nil
}
