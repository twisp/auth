Auth
------

Golang library for exchanging IAM principal for a Twisp OIDC token.

## Usage

Twisp provides an [`http.RoundTripper`](https://pkg.go.dev/net/http#RoundTripper) implementation that handles:

- Exchanging IAM principal for a Twisp OIDC token
- Auto-refreshing those tokens on expiration
- Automatically setting the `Authorization` and `X-Twisp-Account-Id` headers on HTTP requests

```golang
package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/twisp/auth-go"
)

func main() {
    var (
        accountID = "YourAccountID"
        region    = "us-west-2"
        api       = fmt.Sprintf("https://api.%s.cloud.twisp.com/financial/v1/graphql", region)
    )

    client := &http.Client{
        Transport: auth.NewRoundTripper(accountID, region, http.DefaultTransport),
        Timeout:   time.Second * 5,
    }

    query := `{"query": "{ journal { code } }"}`

    req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(query))
    if err != nil {
        log.Fatal(err)
    }

    resp, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    out, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s", string(out))
}
```
