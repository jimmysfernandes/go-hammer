package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HttpClientWrapperJson[T any] struct {
	httpClient      http.Client
	httpClientTrace HttpClientTraceWrapper
	basePath        string
}

func (h HttpClientWrapperJson[T]) Get(ctx context.Context, uri string, extraHeaders map[string]string) (HttpClientWraperResponse[T], error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", h.basePath, strings.TrimPrefix(uri, "/")), nil)
	if err != nil {
		return HttpClientWraperResponse[T]{}, err
	}

	return httpClientDo[T](ctx, &h.httpClient, &h.httpClientTrace, req, extraHeaders)
}

func (h HttpClientWrapperJson[T]) Post(ctx context.Context, uri string, extraHeaders map[string]string, body interface{}) (HttpClientWraperResponse[T], error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return HttpClientWraperResponse[T]{}, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", h.basePath, strings.TrimPrefix(uri, "/")), bytes.NewBuffer(payload))
	if err != nil {
		return HttpClientWraperResponse[T]{}, err
	}

	return httpClientDo[T](ctx, &h.httpClient, &h.httpClientTrace, req, extraHeaders)
}

func NewHttpClientWrapperJson[T any](basePath string) HttpClientWrapperJson[T] {
	basePath = strings.TrimSuffix(basePath, "/")
	basePath = strings.TrimPrefix(basePath, "/")

	httpClient := http.Client{
		Timeout:   30 * time.Second,
		Transport: NewHttpClientTransportJson(),
	}

	httpClientTrace := NewHttpClientTraceWrapper()

	return HttpClientWrapperJson[T]{
		httpClient:      httpClient,
		httpClientTrace: httpClientTrace,
		basePath:        basePath,
	}
}
