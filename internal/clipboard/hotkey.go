package clipboard

import (
	"log"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

type HotkeyTrigger struct {
	stopCh chan struct{}
	WordCh chan string
}

func NewHotkeyTrigger() *HotkeyTrigger {
	return &HotkeyTrigger{
		stopCh: make(chan struct{}),
		WordCh: make(chan string, 1),
	}
}

func (h *HotkeyTrigger) Start() {
	go mainthread.Init(func() {
		hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyW)
		if err := hk.Register(); err != nil {
			log.Printf("failed to register hotkey Ctrl+Shift+W: %v", err)
			return
		}
		log.Println("hotkey registered: Ctrl+Shift+W")

		for {
			select {
			case <-h.stopCh:
				hk.Unregister()
				return
			case <-hk.Keydown():
				if word, ok := ReadClipboardWord(); ok {
					select {
					case h.WordCh <- word:
					default:
					}
				}
			}
		}
	})
}

func (h *HotkeyTrigger) Stop() {
	close(h.stopCh)
}
