package dns

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fsnotify/fsnotify"
)

// resolvConfMonitor monitors /etc/resolv.conf and perfoms actions when the file is changed.
type resolvConfMonitor interface {
	start()
	stop()
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

func (r *resolvConfFileWatcherMonitor) monitorResolvConf(ctx context.Context, doneChan chan<- any) error {
	defer close(doneChan)
	watcher, err := r.getWatcherFunc(resolvconfFilePath)
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer func() {
		if watcher != nil {
			_ = watcher.Close()
		}
	}()

	log.Println(internal.InfoPrefix, dnsPrefix, "starting resolv.conf file watcher")
	for {
		select {
		case e, ok := <-watcher.Events:
			log.Println(internal.InfoPrefix, dnsPrefix, "resolv.conf overwrite detected")
			if !ok {
				return fmt.Errorf("file watcher closed")
			}

			if e.Has(fsnotify.Write | fsnotify.Remove) {
				r.analytics.emitResolvConfOverwrittenEvent()
			}

			return nil
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("file watcher closed")
			}
			log.Println(internal.ErrorPrefix, dnsPrefix, "file watcher error:", err)
		case <-ctx.Done():
			log.Println(internal.InfoPrefix, dnsPrefix, "stopping resolv.conf monitoring")
			return nil
		}
	}
}

// start starts the monitoring goroutine.
func (r *resolvConfFileWatcherMonitor) start() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	r.cancelFunc = cancelFunc
	doneChan := make(chan any)
	r.doneChan = doneChan
	go func() {
		if err := r.monitorResolvConf(ctx, doneChan); err != nil {
			log.Println(internal.ErrorPrefix, dnsPrefix, "resolv.conf monitoring failed:", err)
		}
	}()
}

// stop stops the monitoring goroutine and ensures that it exits before the function return.
func (r *resolvConfFileWatcherMonitor) stop() {
	if r.cancelFunc != nil {
		r.cancelFunc()
		// wait for the monitor goroutine to finish
		select {
		case <-r.doneChan:
		case <-time.After(1 * time.Second):
			log.Println(internal.WarningPrefix, dnsPrefix, "timed out waiting for the monitoring goroutine to stop")
		}
	}
}
