package image

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

type pixabayResponse struct {
	Hits []struct {
		PreviewURL     string `json:"previewURL"`
		WebformatURL   string `json:"webformatURL"`
	} `json:"hits"`
}

func (c *Client) Search(keywords string, count int) ([]string, error) {
	u := fmt.Sprintf("https://pixabay.com/api/?key=%s&q=%s&image_type=photo&per_page=%d&safesearch=true",
		c.apiKey, url.QueryEscape(keywords), count)

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("failed to call Pixabay API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Pixabay API error: %s", string(body))
	}

	var result pixabayResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	urls := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		urls = append(urls, hit.WebformatURL)
	}

	return urls, nil
}
