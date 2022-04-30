package token

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type opt func(*GetTokenOptions)

var OptAssumeRole = func(assumeRoleArn string) opt {
	return func(o *GetTokenOptions) {
		o.AssumeRoleARN = assumeRoleArn
	}
}

// Exchange takes the IAM credentials (from environment) and exchanges
// it for an OIDC token from twisp to use on twisp's /graphql
// endpoint.  The principal on the policies that Twisp evaluates
// should be set to the ARN of the AWS role.
func Exchange(ctx context.Context, account string, region string, opts ...opt) ([]byte, time.Time, error) {
	authURL := fmt.Sprintf("https://auth.%s.%s.twisp.com/", region, account)
	tokenURL := fmt.Sprintf("%stoken/iam", authURL)

	gen, err := NewGenerator(true)
	if err != nil {
		return nil, time.Time{}, err
	}

	o := GetTokenOptions{
		ClusterID: authURL,
		Region:    region,
	}

	for _, op := range opts {
		op(&o)
	}

	t, err := gen.GetWithOptions(&o)
	if err != nil {
		return nil, time.Time{}, err
	}

	j, err := json.Marshal(t)
	if err != nil {
		return nil, time.Time{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewBuffer(j))
	if err != nil {
		return nil, time.Time{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, time.Time{}, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, time.Time{}, fmt.Errorf("iam token exchange got response %d %s", resp.StatusCode, string(body))
	}

	var claims jwt.RegisteredClaims
	if _, _, err := new(jwt.Parser).ParseUnverified(string(body), &claims); err != nil {
		return nil, time.Time{}, err
	}

	return body, claims.ExpiresAt.Time, nil
}
