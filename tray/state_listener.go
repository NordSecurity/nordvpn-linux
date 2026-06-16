package tray

import (
	"context"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
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
		log.Warnf("%s Already listening to daemon events", logTag)
		return
	}

	log.Infof("%s Starting to listen to daemon events", logTag)
	ctx, cancelFunc := context.WithCancel(context.Background())
	l.cancelFunc = cancelFunc

	go l.listen(ctx)
}

func (l *stateListener) Stop() {
	if l.cancelFunc != nil {
		log.Infof("%s Stopping from listening to daemon events", logTag)
		l.cancelFunc()
		l.cancelFunc = nil
	}
}

func (l *stateListener) consumeStream(ctx context.Context, server grpc.ServerStreamingClient[pb.AppState]) {
	for {
		state, err := server.Recv()
		if err != nil {
			log.Errorf("%s Stream receive error: %v", logTag, err)
			return
		}

		select {
		case l.queue <- state:
		case <-time.After(time.Second):
			log.Warnf("%s App state consumer's queue is full, dropping: %v\n", logTag, state)
		case <-ctx.Done():
			return
		}
	}
}

func (l *stateListener) handleAppState(ctx context.Context) {
	for {
		select {
		case item := <-l.queue:
			l.onDataFunc(item)

		case <-ctx.Done():
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

	go l.handleAppState(ctx)

	for {
		if err := RetryWithBackoff(ctx, backoffConfig, op); err != nil {
			log.Infof("%s listen to daemon's state stream: %s\n", logTag, err)
			break
		}

		l.consumeStream(ctx, server)
	}

	close(l.queue)
}
