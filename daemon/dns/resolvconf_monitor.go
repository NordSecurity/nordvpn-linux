package dns

import (
	"context"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fsnotify/fsnotify"
)

type resolvConfMonitor interface {
	Start()
	Stop()
}

type getWatcherFunc func() (*fsnotify.Watcher, error)

type resolvConfFileWatcherMonitor struct {
	analytics      analytics
	getWatcherFunc getWatcherFunc
	cancelFunc     context.CancelFunc
	// doneChan is created when the monitor is started and closed when it is stopped(by calling Done on monitorCtx).
	// It is neccessary to ensure that changes performed on /etc/resolv.conf will not be detected by the monitor when
	// unsetting the DNS.
	doneChan <-chan any
}

func newResolvConfMonitor(analytics analytics) resolvConfFileWatcherMonitor {
	return resolvConfFileWatcherMonitor{
		analytics:      analytics,
		getWatcherFunc: getWatcher,
	}
}

func getWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("starting file watcher: %w", err)
	}

	defer func() {
		if err != nil && watcher != nil {
			_ = watcher.Close()
		}
	}()

	err = watcher.Add(resolvconfFilePath)
	if err != nil {
		return nil, fmt.Errorf("adding file to watchlist: %w", err)
	}

	return watcher, nil
}

func (r *resolvConfFileWatcherMonitor) monitorResolvConf(ctx context.Context) error {
	watcher, err := r.getWatcherFunc()
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer func() {
		if watcher != nil {
			watcher.Close()
		}
	}()

	doneChan := make(chan any)
	r.doneChan = doneChan
	log.Println(internal.InfoPrefix, dnsPrefix, "starting resolv.conf file watcher")
	for {
		select {
		case _, ok := <-watcher.Events:
			log.Println(internal.InfoPrefix, dnsPrefix, "resolv.conf overwrite detected")
			if !ok {
				return fmt.Errorf("file watcher closed")
			}
			r.analytics.emitResolvConfOverwrittenEvent()
			return nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("file watcher closed")
			}
			log.Println(internal.ErrorPrefix, dnsPrefix, "file watcher error:", err)
		case <-ctx.Done():
			close(doneChan)
			log.Println(internal.InfoPrefix, dnsPrefix, "stopping resolv.conf monitoring")
			return nil
		}
	}
}

func (r *resolvConfFileWatcherMonitor) Start() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	r.cancelFunc = cancelFunc
	go func() {
		if err := r.monitorResolvConf(ctx); err != nil {
			log.Println(internal.ErrorPrefix, dnsPrefix, "resolv.conf monitoring failed:", err)
		}
	}()
}

func (r *resolvConfFileWatcherMonitor) Stop() {
	if r.cancelFunc != nil {
		r.cancelFunc()
		// wait for the monitor goroutine to finish
		<-r.doneChan
	}
}
