package main

import (
	"context"
	"fmt"
	"log"

	"create-anki-cards/internal/ai"
	"create-anki-cards/internal/anki"
	"create-anki-cards/internal/cache"
	"create-anki-cards/internal/clipboard"
	"create-anki-cards/internal/config"
	"create-anki-cards/internal/image"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// wordTrigger abstracts clipboard polling vs hotkey trigger.
type wordTrigger interface {
	Start()
	Stop()
}

// imageSearcher abstracts image search providers.
type imageSearcher interface {
	Search(keywords string, count int) ([]string, error)
}

type App struct {
	ctx      context.Context
	config   *config.Config
	ai       *ai.Client
	anki     *anki.Client
	images   []imageSearcher
	cache    *cache.Cache
	trigger  wordTrigger
	wordCh   <-chan string
}

type AnkiCard struct {
	Word     string `json:"word"`
	Meaning  string `json:"meaning"`
	Example  string `json:"example"`
	ImageURL string `json:"image_url"`
}

func NewApp(cfg *config.Config, c *cache.Cache) *App {
	app := &App{
		config: cfg,
		ai:     ai.NewClient(cfg.OpenAIKey),
		anki:   anki.NewClient(cfg.AnkiConnectURL),
		cache:  c,
	}

	if cfg.PixabayAPIKey != "" {
		app.images = append(app.images, image.NewClient(cfg.PixabayAPIKey))
	}
	if cfg.PexelsAPIKey != "" {
		app.images = append(app.images, image.NewPexelsClient(cfg.PexelsAPIKey))
	}

	switch cfg.TriggerMode {
	case "hotkey":
		h := clipboard.NewHotkeyTrigger()
		app.trigger = h
		app.wordCh = h.WordCh
		log.Println("trigger mode: hotkey (Ctrl+Shift+W)")
	default: // "polling"
		m := clipboard.NewMonitor(cfg.PollIntervalMs)
		app.trigger = m
		app.wordCh = m.WordCh
		log.Println("trigger mode: polling")
	}

	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.trigger.Start()
	go a.listenWords()
}

func (a *App) shutdown(ctx context.Context) {
	a.trigger.Stop()
}

// beforeClose intercepts window close and hides instead.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	runtime.WindowHide(ctx)
	return true
}

func (a *App) listenWords() {
	for word := range a.wordCh {
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "word:detected", word)
	}
}

func (a *App) GetCandidates(word string) (*cache.CandidateSet, error) {
	if cs, ok := a.cache.Get(word); ok {
		return cs, nil
	}

	result, err := a.ai.Generate(word)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidates: %w", err)
	}

	synonyms := make([]cache.SynonymEntry, len(result.Synonyms))
	for i, s := range result.Synonyms {
		synonyms[i] = cache.SynonymEntry{Word: s.Word, Difference: s.Difference}
	}

	cs := &cache.CandidateSet{
		Word:          word,
		Meanings:      result.Meanings,
		Examples:      result.Examples,
		CoreImage:     result.CoreImage,
		Synonyms:      synonyms,
		Frequency:     result.Frequency,
		Priority:      result.Priority,
		Usage:         result.Usage,
		ImageKeywords: result.ImageKeywords,
	}

	if err := a.cache.Put(word, cs); err != nil {
		log.Printf("warning: failed to cache result for %q: %v", word, err)
	}

	// Fetch images asynchronously: word itself + first AI-generated keyword
	if len(a.images) > 0 {
		aiKeyword := ""
		if len(result.ImageKeywords) > 0 {
			aiKeyword = result.ImageKeywords[0]
		}
		go a.fetchImages(word, [2]string{word, aiKeyword})
	}

	return cs, nil
}

func (a *App) fetchImages(word string, keywordSets [2]string) {
	// 2 keyword sets × N providers in parallel, 2 images each
	type result struct {
		urls []string
		err  error
	}
	totalJobs := 2 * len(a.images)
	ch := make(chan result, totalJobs)

	for _, kw := range keywordSets {
		if kw == "" {
			totalJobs -= len(a.images)
			continue
		}
		for _, provider := range a.images {
			go func(p imageSearcher, keywords string) {
				urls, err := p.Search(keywords, 3)
				if len(urls) > 2 {
					urls = urls[:2]
				}
				ch <- result{urls, err}
			}(provider, kw)
		}
	}

	var allURLs []string
	for i := 0; i < totalJobs; i++ {
		r := <-ch
		if r.err != nil {
			log.Printf("warning: image search failed for %q: %v", word, r.err)
			continue
		}
		allURLs = append(allURLs, r.urls...)
	}

	if len(allURLs) == 0 {
		return
	}

	// Update cache with image URLs
	if cs, ok := a.cache.Get(word); ok {
		cs.ImageURLs = allURLs
		if err := a.cache.Put(word, cs); err != nil {
			log.Printf("warning: failed to update cache for %q: %v", word, err)
		}
	}

	// Notify frontend
	runtime.EventsEmit(a.ctx, "images:loaded", map[string]interface{}{
		"word":       word,
		"image_urls": allURLs,
	})
}

func (a *App) RegisterToAnki(card AnkiCard) error {
	fields := make(map[string]string)
	for key, fieldName := range a.config.FieldMapping {
		switch key {
		case "word":
			fields[fieldName] = card.Word
		case "meaning":
			fields[fieldName] = card.Meaning
		case "example":
			fields[fieldName] = card.Example
		}
	}

	imageField := a.config.FieldMapping["image"]
	return a.anki.AddNote(a.config.DeckName, a.config.NoteType, fields, card.ImageURL, imageField)
}

func (a *App) CheckAnkiConnection() (bool, error) {
	err := a.anki.CheckConnection()
	if err != nil {
		return false, err
	}
	return true, nil
}

// HideWindow hides the window (called from frontend after dismiss).
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

// QuitApp exits the application.
func (a *App) QuitApp() {
	runtime.Quit(a.ctx)
}
