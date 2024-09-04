package sharedctx

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// New is a default constructor for Context.
func New() *Context {
	return &Context{
		constructorFn: func() (context.Context, context.CancelFunc) {
			return context.WithTimeout(context.Background(), time.Second*30)
		},
	}
}

// Context can be used to not allow multiple parties to execute similar action simultaneously
// without unifying the action executor (e. g. VPN connect and Meshnet exit node connect).
type Context struct {
	mu            sync.Mutex
	execMu        sync.Mutex
	ctx           context.Context
	cancelFunc    *cancelFunc
	constructorFn func() (context.Context, context.CancelFunc)
}

// cancelFunc is a stateful wrapper around `context.CancelFunc` which stores whether an internal
// function was called or not.
type cancelFunc struct {
	f      atomic.Pointer[context.CancelFunc]
	called atomic.Bool
}

// CancelFunc returns the cancel function of a currently executed function. If nothing is being
// executed at the moment, it returns nil.
//
// Thread safe.
func (c *Context) CancelFunc() (context.CancelFunc, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelFunc == nil {
		return nil, false
	}
	return c.cancelFunc.call, c.cancelFunc.called.Load()
}

// TryExecuteWith executes the given function with a new context provided to it. If another action
// is currently being executed, it will return `false` immediately. Otherwise, it will block until
// given function is executed and will return `true`
//
// Thread safe.
func (c *Context) TryExecuteWith(f func(context.Context)) bool {
	if !c.execMu.TryLock() {
		return false
	}
	c.mu.Lock()
	ctx, cancel := c.constructorFn()
	defer cancel()
	c.ctx = ctx
	c.cancelFunc = &cancelFunc{}
	c.cancelFunc.f.Store(&cancel)
	c.mu.Unlock()
	f(c.ctx)
	c.mu.Lock()
	c.ctx = nil
	c.cancelFunc = nil
	c.mu.Unlock()
	c.execMu.Unlock()
	return true
}

func (cf *cancelFunc) call() {
	f := cf.f.Load()
	if f == nil {
		return
	}
	(*f)()
	cf.called.Store(true)
}
