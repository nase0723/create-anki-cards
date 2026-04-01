package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"create-anki-cards/internal/cache"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

type aiResult struct {
	Meanings      []string        `json:"meanings"`
	Examples      []string        `json:"examples"`
	CoreImage     string          `json:"core_image"`
	RawSynonyms   json.RawMessage `json:"synonyms"`
	Frequency     string          `json:"frequency"`
	Priority      string          `json:"priority"`
	Usage         string          `json:"usage"`
	ImageKeywords []string        `json:"image_keywords"`
}

func (r *aiResult) parseSynonyms() []cache.SynonymEntry {
	var entries []cache.SynonymEntry
	if err := json.Unmarshal(r.RawSynonyms, &entries); err == nil {
		return entries
	}
	var strs []string
	if err := json.Unmarshal(r.RawSynonyms, &strs); err == nil {
		for _, s := range strs {
			entries = append(entries, cache.SynonymEntry{Word: s})
		}
	}
	return entries
}

func (r *aiResult) toCandidateSet(word string) *cache.CandidateSet {
	return &cache.CandidateSet{
		Word:          word,
		Meanings:      r.Meanings,
		Examples:      r.Examples,
		CoreImage:     r.CoreImage,
		Synonyms:      r.parseSynonyms(),
		Frequency:     r.Frequency,
		Priority:      r.Priority,
		Usage:         r.Usage,
		ImageKeywords: r.ImageKeywords,
	}
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *respFormat   `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type respFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) chatCompletion(prompt string) (string, error) {
	req := chatRequest{
		Model: "gpt-4o-mini",
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &respFormat{Type: "json_object"},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *Client) Generate(word string) (*cache.CandidateSet, error) {
	prompt := fmt.Sprintf(`You are an English vocabulary assistant. Given the English word "%s", provide:
1. Three different Japanese meanings (concise, dictionary-style)
2. Three short example sentences using the word (under 10 words each, natural English)
3. Core image in Japanese: one short sentence capturing the word's essential feeling or concept (e.g. "「押しのける」感覚" for "push")
4. Three commonly used synonyms with how each differs from "%s" (difference in Japanese, concise)
5. Frequency level: how common this word is in everyday English (one of: "Very Common", "Common", "Uncommon", "Rare")
6. Anki priority: how worthwhile it is for a learner to memorize this word (one of: "High", "Medium", "Low")
7. Usage context in Japanese: where and when this word is commonly used (e.g. "ビジネスメールや論文でよく使われるフォーマルな表現")
8. image_keywords: MUST be a JSON array of exactly 2 strings, each 2-3 concrete visual English words from different angles for photo search (do NOT include the word itself). Example for "integrity": ["handshake agreement", "shield honor"]

Respond in JSON format:
{"meanings": ["意味1", "意味2", "意味3"], "examples": ["sentence1", "sentence2", "sentence3"], "core_image": "コアイメージ", "synonyms": [{"word": "synonym1", "difference": "違いの説明"}], "frequency": "Common", "priority": "High", "usage": "使用場面の説明", "image_keywords": ["visual angle 1", "visual angle 2"]}`, word, word)

	content, err := c.chatCompletion(prompt)
	if err != nil {
		return nil, err
	}

	var result aiResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI result: %w", err)
	}

	return result.toCandidateSet(word), nil
}

func (c *Client) GenerateImageKeywords(word string, excludeKeywords []string) (string, error) {
	prompt := fmt.Sprintf(`For the English word "%s", generate 2-3 concrete visual English words for photo search.
Do NOT use any of these previously used keywords: %s
Respond in JSON: {"keywords": "new visual words"}`, word, strings.Join(excludeKeywords, ", "))

	content, err := c.chatCompletion(prompt)
	if err != nil {
		return "", err
	}

	var result struct {
		Keywords string `json:"keywords"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return "", fmt.Errorf("failed to parse keywords: %w", err)
	}

	return result.Keywords, nil
}
