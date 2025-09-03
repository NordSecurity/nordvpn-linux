package tray

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
)

type stateListener struct {
	client     pb.DaemonClient
	queue      chan *pb.AppState
	cancelFunc context.CancelFunc
	onDataFunc func(item *pb.AppState)
}

func newStateListener(client pb.DaemonClient, onDataFunc func(item *pb.AppState)) *stateListener {
	return &stateListener{
		client:     client,
		queue:      make(chan *pb.AppState, 10),
		onDataFunc: onDataFunc,
	}
}

func (l *stateListener) Start() {
	if l.cancelFunc != nil {
		log.Printf("%s %s Already listening to daemon events", logTag, internal.WarningPrefix)
		return
	}

	log.Printf("%s %s Starting to listen to daemon events", logTag, internal.InfoPrefix)
	ctx, cancelFunc := context.WithCancel(context.Background())
	l.cancelFunc = cancelFunc

	go l.listen(ctx)
}

func (l *stateListener) Stop() {
	if l.cancelFunc != nil {
		log.Printf("%s %s Stopping from listening to daemon events", logTag, internal.InfoPrefix)
		l.cancelFunc()
		l.cancelFunc = nil
	}
}

func (l *stateListener) consumeStream(server grpc.ServerStreamingClient[pb.AppState]) {
	for {
		state, err := server.Recv()
		if err != nil {
			log.Printf("%s %s Stream receive error: %v\n", logTag, internal.ErrorPrefix, err)
			if strings.Contains(err.Error(), "EOF") {
				l.cancelFunc()
			}
			return
		}

		select {
		case l.queue <- state:
		case <-time.After(time.Second):
			log.Printf("%s %s App state consumer's queue is full, dropping", logTag, internal.WarningPrefix)
		}
	}
}

func (c *stateListener) handleAppState(ctx context.Context) {
	for {
		select {
		case item := <-c.queue:
			c.onDataFunc(item)

		case <-ctx.Done():
			log.Printf("%s %s exiting systray\n", logTag, internal.InfoPrefix)
			defer close(c.queue)
			return
		}
	}
}

func (l *stateListener) listen(ctx context.Context) {
	var server grpc.ServerStreamingClient[pb.AppState]

	// setup operation retry mechanism to retry indefinitely
	backoffConfig := BackoffConfig{
		MaxDelay: 10 * time.Second,
	}
	op := func(backoffCtx context.Context) error {
		svr, err := l.client.SubscribeToStateChanges(backoffCtx, nil)
		if err == nil {
			server = svr
		}
		return err
	}
	if err := RetryWithBackoff(ctx, backoffConfig, op); err != nil {
		log.Printf("%s %s listen to daemon's state stream: %s\n", logTag, internal.ErrorPrefix, err)
		return
	}

	go l.handleAppState(ctx)

	l.consumeStream(server)
}
