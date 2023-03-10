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

type Loader struct {
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
	return Loader{
		active:   false,
		stopChan: make(chan struct{}, 1),
		lock:     &sync.RWMutex{},
	}
}

func (l *Loader) Start() {
	l.StartWithConfig(nil)
}

func (l *Loader) StartWithConfig(config *LoaderConfig) {
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

func (l *Loader) Stop() {
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

func (l *Loader) IsActive() bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.active
}
