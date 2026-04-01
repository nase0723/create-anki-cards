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

type wordTrigger interface {
	Start()
	Stop()
}

type imageSearcher interface {
	Search(keywords string, count int) ([]string, error)
}

const maxImages = 9

type App struct {
	ctx     context.Context
	config  *config.Config
	ai      *ai.Client
	anki    *anki.Client
	images  []imageSearcher
	cache   *cache.Cache
	trigger wordTrigger
	wordCh  <-chan string
}

type AnkiCard struct {
	Word      string   `json:"word"`
	Meaning   string   `json:"meaning"`
	Example   string   `json:"example"`
	ImageURLs []string `json:"image_urls"`
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
	default:
		m := clipboard.NewMonitor(cfg.PollIntervalMs)
		app.trigger = m
		app.wordCh = m.WordCh
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
	isDuplicate, _ := a.CheckDuplicate(word)

	if cs, ok := a.cache.Get(word); ok {
		cs.IsDuplicate = isDuplicate
		return cs, nil
	}

	cs, err := a.ai.Generate(word)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidates: %w", err)
	}
	cs.IsDuplicate = isDuplicate

	if err := a.cache.Put(word, cs); err != nil {
		log.Printf("warning: failed to cache result for %q: %v", word, err)
	}

	if len(a.images) > 0 {
		keywords := append([]string{word}, cs.ImageKeywords...)
		go a.emitImages(word, a.searchImages(keywords, nil))
	}

	return cs, nil
}

// searchImages fetches images from all providers in parallel, excluding previousURLs.
func (a *App) searchImages(keywords []string, previousURLs map[string]bool) []string {
	type result struct {
		urls []string
		err  error
	}

	var validKeywords []string
	for _, kw := range keywords {
		if kw != "" {
			validKeywords = append(validKeywords, kw)
		}
	}

	totalJobs := len(validKeywords) * len(a.images)
	ch := make(chan result, totalJobs)

	for _, kw := range validKeywords {
		for _, provider := range a.images {
			go func(p imageSearcher, keywords string) {
				urls, err := p.Search(keywords, 5)
				ch <- result{urls, err}
			}(provider, kw)
		}
	}

	var allURLs []string
	for i := 0; i < totalJobs; i++ {
		r := <-ch
		if r.err != nil {
			log.Printf("warning: image search failed: %v", r.err)
			continue
		}
		for _, u := range r.urls {
			if previousURLs == nil || !previousURLs[u] {
				allURLs = append(allURLs, u)
			}
		}
	}

	if len(allURLs) > maxImages {
		allURLs = allURLs[:maxImages]
	}
	return allURLs
}

// emitImages updates cache and notifies frontend.
func (a *App) emitImages(word string, urls []string) {
	if len(urls) == 0 {
		return
	}
	if cs, ok := a.cache.Get(word); ok {
		cs.ImageURLs = urls
		_ = a.cache.Put(word, cs)
	}
	runtime.EventsEmit(a.ctx, "images:loaded", map[string]interface{}{
		"word":       word,
		"image_urls": urls,
	})
}

func (a *App) RefreshImages(word string) {
	if len(a.images) == 0 {
		return
	}

	var excludeKeywords []string
	if cs, ok := a.cache.Get(word); ok {
		excludeKeywords = cs.ImageKeywords
	}

	newKeyword, err := a.ai.GenerateImageKeywords(word, excludeKeywords)
	if err != nil {
		log.Printf("warning: failed to generate new keywords for %q: %v", word, err)
		return
	}

	if cs, ok := a.cache.Get(word); ok {
		cs.ImageKeywords = append(cs.ImageKeywords, newKeyword)
		_ = a.cache.Put(word, cs)
	}

	urls := a.searchImages([]string{newKeyword}, nil)
	if len(urls) == 0 {
		return
	}

	if cs, ok := a.cache.Get(word); ok {
		cs.ImageURLs = urls
		_ = a.cache.Put(word, cs)
	}
	runtime.EventsEmit(a.ctx, "images:refreshed", map[string]interface{}{
		"word":       word,
		"image_urls": urls,
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
	return a.anki.AddNote(a.config.DeckName, a.config.NoteType, fields, card.ImageURLs, a.config.FieldMapping["image"])
}

func (a *App) CheckDuplicate(word string) (bool, error) {
	wordField := a.config.FieldMapping["word"]
	ids, err := a.anki.FindNotes(fmt.Sprintf("%s:%s", wordField, word))
	if err != nil {
		return false, err
	}
	return len(ids) > 0, nil
}

func (a *App) BrowseInAnki(word string) error {
	wordField := a.config.FieldMapping["word"]
	return a.anki.GuiBrowse(fmt.Sprintf("%s:%s", wordField, word))
}

func (a *App) CheckAnkiConnection() (bool, error) {
	if err := a.anki.CheckConnection(); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

func (a *App) QuitApp() {
	runtime.Quit(a.ctx)
}
