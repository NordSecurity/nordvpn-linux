package cli

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// loader defines loader's every frame
var loader = []string{
	"-",
	"\\",
	"|",
	"/",
}

// Loader is show progress interface
type Loader interface {
	Start()
	Stop()
}

// SilentLoader do not show progress in case of non-terminal output
type SilentLoader struct{}

func (l *SilentLoader) Start() {}
func (l *SilentLoader) Stop()  {}

// TerminalLoader show operation progress to the human user in the terminal
type TerminalLoader struct {
	lastWrite int
	active    bool
	stopChan  chan struct{}
	lock      *sync.RWMutex
	config    *LoaderConfig
}

type LoaderConfig struct {
	Prefix string
}

func NewLoader() Loader {
	if isStdoutATerminal() {
		return &TerminalLoader{
			active:   false,
			stopChan: make(chan struct{}, 1),
			lock:     &sync.RWMutex{},
		}
	}
	return &SilentLoader{}
}

func (l *TerminalLoader) Start() {
	l.startWithConfig(nil)
}

func (l *TerminalLoader) startWithConfig(config *LoaderConfig) {
	l.lock.Lock()
	if l.active {
		l.lock.Unlock()
		return
	}
	l.active = true
	l.config = config
	l.lock.Unlock()
	go func() {
		for {
			for _, c := range loader {
				select {
				case <-l.stopChan:
					return
				default:
					l.lock.Lock()
					if l.config != nil {
						l.lastWrite, _ = fmt.Printf("\r%s %s", l.config.Prefix, c)
					} else {
						l.lastWrite, _ = fmt.Printf("\r%s", c)
					}
					l.lock.Unlock()
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
	}()
}

func (l *TerminalLoader) Stop() {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.active {
		l.active = false
		if l.config != nil {
			fmt.Printf("\r%s\r%s %s\n", strings.Repeat(" ", l.lastWrite), l.config.Prefix, color.GreenString("DONE"))
		} else {
			fmt.Printf("\r%s\r", strings.Repeat(" ", l.lastWrite))
		}
		l.stopChan <- struct{}{}
	}
	l.config = nil
}
