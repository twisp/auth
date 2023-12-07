package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/twisp/auth/token"
	"golang.org/x/sync/singleflight"
)

var (
	expired = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func NewTwispDefaultRoundTripper(customerAccount string, region string) http.RoundTripper {
	return NewTwispRoundTripper(customerAccount, "cloud", region, time.Now)
}

func NewTwispRoundTripper(customerAccount, twispEnvironment, region string, now Now) http.RoundTripper {
	return &roundTripper{
		customerAccount:  customerAccount,
		twispEnvironment: twispEnvironment,
		region:           region,
		now:              now,
		expire:           expired,
		single:           new(singleflight.Group),
		auth:             []byte{},
		wrapped:          http.DefaultTransport,
	}
}

type roundTripper struct {
	customerAccount  string
	twispEnvironment string
	region           string

	now    func() time.Time
	auth   []byte
	expire time.Time

	single *singleflight.Group

	wrapped http.RoundTripper
}

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	jwt, err := r.authorization()
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", string(jwt)))
	req.Header.Set("X-Twisp-Account-Id", r.customerAccount)

	return r.wrapped.RoundTrip(req)
}

// authorization returns an OIDC token from Twisp by exchanging the current IAM credentials.
// The OIDC token is cached until it expires.
func (t *roundTripper) authorization() ([]byte, error) {
	// Read cached version
	if len(t.auth) > 0 && t.now().Before(t.expire) {
		return t.auth, nil
	}

	// We need to get a new token.
	auth, err, _ := t.single.Do("authorization", func() (any, error) {
		// Double check
		if len(t.auth) > 0 && t.now().Before(t.expire) {
			return t.auth, nil
		}

		b, err := token.Exchange(t.twispEnvironment, t.region)
		if err != nil {
			return nil, err
		}
		exp, err := extractExpire(string(b))
		if err != nil {
			return nil, err
		}

		t.auth = b
		t.expire = exp
		return b, nil
	})
	if err != nil {
		return nil, err
	}

	return auth.([]byte), nil
}

// extractExpire extracts the exp claim from the jwt and returns it.
func extractExpire(jwt string) (time.Time, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return expired, fmt.Errorf("verify: malformed jwt, expected 3 parts got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return expired, fmt.Errorf("verify: malformed jwt payload: %v", err)
	}

	type idToken struct {
		Expire int64 `json:"exp"`
	}

	var token idToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return expired, fmt.Errorf("verify: failed to unmarshal claims: %v", err)
	}

	return time.Unix(token.Expire, 0), nil
}

type Now func() time.Time
