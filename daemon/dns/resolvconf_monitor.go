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

type getWatcherFunc func(pathsToMonitor ...string) (*fsnotify.Watcher, error)

type resolvConfFileWatcherMonitor struct {
	analytics      analytics
	getWatcherFunc getWatcherFunc
	cancelFunc     context.CancelFunc
	// doneChan is created when the monitor is started and closed when it is stopped(by calling Done on monitorCtx).
	// It is necessary to ensure that changes performed on /etc/resolv.conf will not be detected by the monitor when
	// unsetting the DNS.
	doneChan <-chan any
}

func newResolvConfMonitor(analytics analytics) resolvConfFileWatcherMonitor {
	return resolvConfFileWatcherMonitor{
		analytics:      analytics,
		getWatcherFunc: internal.GetFileWatcher,
	}
}

func (r *resolvConfFileWatcherMonitor) monitorResolvConf(ctx context.Context) error {
	watcher, err := r.getWatcherFunc(resolvconfFilePath)
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer func() {
		if watcher != nil {
			_ = watcher.Close()
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
