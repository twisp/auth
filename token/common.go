package token

import (
	"context"
	"time"
)

type TokenGenerator interface {
	Generate(context.Context) ([]byte, time.Time, error)
}

type TokenRefresher interface {
	Token() ([]byte, bool, error)
	Stop()
}
