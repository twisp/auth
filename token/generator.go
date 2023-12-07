package token

import (
	"context"
	"time"
)

type tokenGeneratorStatic struct {
	value []byte
}

var _ TokenGenerator = (*tokenGeneratorStatic)(nil)

func NewTokenGeneratorStatic(value []byte) TokenGenerator {
	return &tokenGeneratorStatic{
		value: value,
	}
}

func (g *tokenGeneratorStatic) Generate(_ context.Context) ([]byte, time.Time, error) {
	return g.value, time.Unix(1<<63-1, 0), nil
}

type tokenGeneratorIAM struct {
	account string
	region  string
}

var _ TokenGenerator = (*tokenGeneratorIAM)(nil)

func NewTokenGeneratorIAM(account, region string) TokenGenerator {
	return &tokenGeneratorIAM{
		account: account,
		region:  region,
	}
}

func (g *tokenGeneratorIAM) Generate(ctx context.Context) ([]byte, time.Time, error) {
	return Exchange(ctx, g.account, g.region)
}
