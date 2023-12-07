Auth
------

Golang helper library for exchanging IAM principal for a Twisp OIDC token.

## Example

Twisp provides an [`http.RoundTripper`](https://pkg.go.dev/net/http#RoundTripper) implementation that handles:

- Exchanging IAM principal for Twisp OIDC
- Setting bearer token and Twisp account id on headers
- Refreshing token when expired

