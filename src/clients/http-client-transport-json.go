package clients

import (
	"net/http"

	"github.com/google/uuid"
)

type httpClientTransportJson struct {
	T http.RoundTripper
}

func (a *httpClientTransportJson) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("X-Idepotentid", uuid.NewString())
	return a.T.RoundTrip(req)
}

func NewHttpClientTransportJson() http.RoundTripper {
	return &httpClientTransportJson{
		T: &http.Transport{
			DisableKeepAlives: true,
		},
	}
}
