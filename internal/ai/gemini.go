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
	model  string
}

func NewClient(apiKey, model string) *Client {
	return &Client{apiKey: apiKey, model: model}
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

type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig *geminiGenConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) chatCompletion(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)

	req := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: &geminiGenConfig{
			ResponseMimeType: "application/json",
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
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
