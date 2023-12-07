package token

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type tokenRefresherAlways struct {
	generator TokenGenerator
}

func NewTokenRefresherAlways(generator TokenGenerator) TokenRefresher {
	return &tokenRefresherAlways{
		generator: generator,
	}
}

func (r *tokenRefresherAlways) Token() ([]byte, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	generated, _, err := r.generator.Generate(ctx)
	cancel()
	return generated, true, err
}

func (r *tokenRefresherAlways) Stop() {}

type tokenRefresherTTL struct {
	generator     TokenGenerator
	maxTokenAge   time.Duration
	maxRefreshAge time.Duration
	now           func() time.Time
	token         atomic.Pointer[tokenWithTTL]
	mux           sync.Mutex
	refreshT      *time.Timer
	refreshStopC  chan struct{}
	stopped       bool
}

var _ TokenRefresher = (*tokenRefresherTTL)(nil)

type tokenWithTTL struct {
	token []byte
	exp   time.Time
	now   func() time.Time
}

func (t *tokenWithTTL) Valid() bool {
	if t == nil {
		return false
	}
	return t.exp.After(t.now())
}

func NewTokenRefresherTTL(
	generator TokenGenerator,
	maxTokenAge time.Duration,
	maxRefreshAge time.Duration,
	now func() time.Time,
) TokenRefresher {
	r := &tokenRefresherTTL{
		generator:     generator,
		maxTokenAge:   maxTokenAge,
		maxRefreshAge: maxRefreshAge,
		now:           now,
		refreshT:      time.NewTimer(time.Minute),
		refreshStopC:  make(chan struct{}),
		stopped:       false,
	}

	r.refreshT.Stop()

	go func() {
		for {
			select {
			case <-r.refreshStopC:
				return
			case <-r.refreshT.C:
				_, _, _ = r.refresh()
			}
		}
	}()

	// pre-emptively load our first token

	go func() {
		_, _, _ = r.Token()
	}()

	return r
}

func (r *tokenRefresherTTL) Token() ([]byte, bool, error) {
	if token := r.token.Load(); token.Valid() {
		return token.token, false, nil
	}
	return r.refresh()
}

func (r *tokenRefresherTTL) refresh() ([]byte, bool, error) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if token := r.token.Load(); token.Valid() {
		return token.token, false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, exp, err := r.generator.Generate(ctx)
	if err != nil {
		return nil, false, err
	}

	r.token.Store(&tokenWithTTL{
		token: token,
		exp:   exp.Add(-r.maxTokenAge),
		now:   r.now,
	})

	if !r.stopped {
		refreshIn := exp.Sub(r.now().Add(-r.maxRefreshAge))
		if !r.refreshT.Stop() {
			select {
			case <-r.refreshT.C:
			default:
			}
		}
		r.refreshT.Reset(refreshIn)
	}

	return token, true, nil
}

func (r *tokenRefresherTTL) Stop() {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.stopped = true

	close(r.refreshStopC)

	if !r.refreshT.Stop() {
		select {
		case <-r.refreshT.C:
		default:
		}
	}
}
