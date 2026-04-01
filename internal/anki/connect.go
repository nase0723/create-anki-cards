package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

type ankiRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

type ankiResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *string         `json:"error"`
}

func (c *Client) CheckConnection() error {
	_, err := c.invoke("version", nil)
	return err
}

type addNoteParams struct {
	Note noteParams `json:"note"`
}

type noteParams struct {
	DeckName  string                 `json:"deckName"`
	ModelName string                 `json:"modelName"`
	Fields    map[string]string      `json:"fields"`
	Options   map[string]interface{} `json:"options"`
	Picture   []notePicture          `json:"picture,omitempty"`
}

type notePicture struct {
	URL      string   `json:"url"`
	Filename string   `json:"filename"`
	Fields   []string `json:"fields"`
}

func (c *Client) AddNote(deckName, noteType string, fields map[string]string, imageURL string, imageField string) error {
	note := noteParams{
		DeckName:  deckName,
		ModelName: noteType,
		Fields:    fields,
		Options: map[string]interface{}{
			"allowDuplicate": false,
		},
	}

	if imageURL != "" && imageField != "" {
		note.Picture = []notePicture{
			{
				URL:      imageURL,
				Filename: fields[imageField] + ".jpg",
				Fields:   []string{imageField},
			},
		}
	}

	params := addNoteParams{Note: note}
	_, err := c.invoke("addNote", params)
	return err
}

func (c *Client) invoke(action string, params interface{}) (json.RawMessage, error) {
	req := ankiRequest{
		Action:  action,
		Version: 6,
		Params:  params,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AnkiConnect: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var ankiResp ankiResponse
	if err := json.Unmarshal(respBody, &ankiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ankiResp.Error != nil {
		return nil, fmt.Errorf("AnkiConnect error: %s", *ankiResp.Error)
	}

	return ankiResp.Result, nil
}
