package signal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Handler struct {
	killFunc    func()
	exitFunc    func()
	firstSignal time.Time
	signalCount int
	mutex       sync.Mutex
}

func NewHandler(killFunc, exitFunc func()) *Handler {
	return &Handler{
		killFunc: killFunc,
		exitFunc: exitFunc,
	}
}

func (h *Handler) Start() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	for range sigChan {
		h.handleSignal()
	}
}

func (h *Handler) handleSignal() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	now := time.Now()
	
	if h.signalCount == 0 {
		h.firstSignal = now
		h.signalCount = 1
		h.killFunc()
	} else {
		if now.Sub(h.firstSignal) <= time.Second {
			h.exitFunc()
		} else {
			h.firstSignal = now
			h.signalCount = 1
			h.killFunc()
		}
	}
}