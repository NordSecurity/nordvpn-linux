package sharedctx

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestContext_TryExecuteWith(t *testing.T) {
	category.Set(t, category.Unit)
	longFn := func(sctx *Context, ch chan<- struct{}) {
		sctx.TryExecuteWith(func(ctx context.Context) {
			ch <- struct{}{}
			<-ctx.Done()
			ch <- struct{}{}
		})
	}

	testExecute := func(ctx *Context, success bool) {
		for i := 0; i < 2; i++ {
			executed := false
			assert.Equal(t,
				success,
				ctx.TryExecuteWith(func(context.Context) { executed = true }))
			assert.Equal(t, success, executed)
			cancelFn, canceled := ctx.CancelFunc()
			assert.False(t, canceled)
			assert.Equal(t, success, cancelFn == nil)

			assert.Equal(t, success, ctx.cancelFunc == nil)
			if !success {
				assert.NotNil(t, success, ctx.cancelFunc.f.Load())
			}
			assert.Equal(t, success, ctx.ctx == nil)
		}
	}

	ctx := New()

	// Function should execute immediately multiple times in a row as nothing is blocking it
	testExecute(ctx, true)

	// Function shoud immediately return false if it is blocked
	ch := make(chan struct{}, 1)
	defer close(ch)
	go longFn(ctx, ch)
	// Make sure longFn starts before testExecute
	<-ch
	testExecute(ctx, false)

	// After long function is done, context should be clean again
	cancelFn, _ := ctx.CancelFunc()
	cancelFn()
	// Make sure longFn exits before testExecute as context.CancelFunc does not wait for
	// anything
	<-ch
	testExecute(ctx, true)
}

func TestContext_CancelFunc(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name          string
		cancelFuncNil bool
		called        bool
		setup         func() *Context
	}{
		{
			name:          "Initial ctx",
			cancelFuncNil: true,
			called:        false,
			setup:         New,
		},
		{
			name:   "cancelFunc not nil but empty",
			called: false,
			setup: func() *Context {
				ctx := New()
				ctx.cancelFunc = &cancelFunc{}
				return ctx
			},
		},
		{
			name:   "cancelFunc not nil",
			called: true,
			setup: func() *Context {
				ctx := New()
				ctx.cancelFunc = &cancelFunc{}
				f := context.CancelFunc(func() {})
				ctx.cancelFunc.f.Store(&f)
				return ctx
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			cancelFunc, called := ctx.CancelFunc()
			assert.False(t, called)
			assert.Equal(t, tt.cancelFuncNil, cancelFunc == nil)
			if cancelFunc != nil {
				cancelFunc()
			}
			cancelFunc, called = ctx.CancelFunc()
			assert.Equal(t, tt.cancelFuncNil, cancelFunc == nil)
			assert.Equal(t, tt.called, called)
		})
	}
}

func TestCancelFunc_Call(t *testing.T) {
	category.Set(t, category.Unit)
	called := false
	tests := []struct {
		name   string
		called bool
		f      context.CancelFunc
	}{
		{
			name:   "Cancel function is nil",
			called: false,
			f:      nil,
		},
		{
			name:   "Cancel function is non-nil",
			called: true,
			f:      func() { called = true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cf cancelFunc
			if tt.f != nil {
				cf.f.Store(&tt.f)
			}
			// Call the method
			cf.call()
			assert.Equal(t, tt.called, cf.called.Load())
			assert.Equal(t, tt.called, called)
		})
		called = false
	}
}
