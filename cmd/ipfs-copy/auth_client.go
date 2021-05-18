package main

import "net/http"

// NewClient create an http.Client that automatically perform basic auth on each requests.
//
// This is a bit hacky but also the only way I found to have basic authentication with the
// normal IPFS API wrapper that only allow the usage of a custom http.Client.
// Here, we build a custom Client with a custom transport that apply the basic auth on each request.
func NewClient(username, password string) *http.Client {
	return &http.Client{
		Transport: authTransport{
			RoundTripper: http.DefaultTransport,
			Username:     username,
			Password:     password,
		},
	}
}

// authTransport is a transport that also apply basic auth header on each requests made
type authTransport struct {
	http.RoundTripper
	Username string
	Password string
}

func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.Username, t.Password)
	return t.RoundTripper.RoundTrip(r)
}
