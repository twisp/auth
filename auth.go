package auth

import (
	"net/http"
	"time"

	"github.com/twisp/auth-go/token"
)

var (
	DefaultClock         = time.Now
	DefaultMaxTokenAge   = time.Minute
	DefaultMaxRefreshAge = 5 * time.Minute
)

const (
	DefaultEnv          = "cloud"
	HeaderAuthorization = "Authorization"
	HeaderAccountID     = "X-Twisp-Account-Id"
)

type authRoundTripper struct {
	accountID string
	refresher token.TokenRefresher
	transport http.RoundTripper
}

func NewRoundTripper(accountID string, region string, transport http.RoundTripper) http.RoundTripper {
	return NewEnvironmentRoundTripper(accountID, DefaultEnv, region, transport)
}

func NewEnvironmentRoundTripper(accountID, env, region string, transport http.RoundTripper) http.RoundTripper {
	return &authRoundTripper{
		accountID: accountID,
		refresher: token.NewTokenRefresherTTL(
			token.NewTokenGeneratorIAM(env, region),
			DefaultMaxTokenAge,
			DefaultMaxRefreshAge,
			DefaultClock,
		),
		transport: transport,
	}
}

func (r *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	jwt, _, err := r.refresher.Token()
	if err != nil {
		return nil, err
	}

	req.Header.Set(HeaderAuthorization, "Bearer "+string(jwt))
	req.Header.Set(HeaderAccountID, r.accountID)

	return r.transport.RoundTrip(req)
}
