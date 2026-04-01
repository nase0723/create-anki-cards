package clipboard

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
)

var wordPattern = regexp.MustCompile(`^[a-zA-Z]{2,30}$`)

type Monitor struct {
	interval time.Duration
	lastText string
	cancel   context.CancelFunc
	WordCh   chan string
}

func NewMonitor(intervalMs int) *Monitor {
	return &Monitor{
		interval: time.Duration(intervalMs) * time.Millisecond,
		WordCh:   make(chan string, 1),
	}
}

func (m *Monitor) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	go m.poll(ctx)
}

func (m *Monitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Monitor) poll(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	m.lastText, _ = clipboard.ReadAll()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			text, err := clipboard.ReadAll()
			if err != nil {
				continue
			}

			text = strings.TrimSpace(text)
			if text == m.lastText {
				continue
			}
			m.lastText = text

			if wordPattern.MatchString(text) {
				select {
				case m.WordCh <- strings.ToLower(text):
				default:
				}
			}
		}
	}
}

// ReadClipboardWord reads the current clipboard and returns the word if valid.
func ReadClipboardWord() (string, bool) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", false
	}
	text = strings.TrimSpace(text)
	if wordPattern.MatchString(text) {
		return strings.ToLower(text), true
	}
	return "", false
}
