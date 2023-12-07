package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//Exchange takes the IAM credentials (from environment) and exchanges it for an OIDC
//token from twisp to use on twisp's /graphql endpoint.  The principal on the policies
//that Twisp evaluates should be set to the ARN of the AWS role.
func Exchange(account string, region string) ([]byte, error) {
	authURL := fmt.Sprintf("https://auth.%s.%s.twisp.com/", region, account)
	tokenURL := fmt.Sprintf("%stoken/iam", authURL)

	gen, err := NewGenerator(true)
	if err != nil {
		return nil, err
	}

	o := GetTokenOptions{
		ClusterID: authURL,
		Region:    region,
	}

	t, err := gen.GetWithOptions(&o)
	if err != nil {
		return nil, err
	}

	j, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("iam token exchange got response %d %s", resp.StatusCode, string(body))
	}

	return body, nil
}
