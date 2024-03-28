package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpClientWraperResponse[T any] struct {
	Body       T      `json:"body"`
	StatusCode int    `json:"status_code"`
	Status     string `json:"status"`
	Err        error  `json:"error"`
}

func newHttpClientWraperResponseWithoutResponse[T any](err error) HttpClientWraperResponse[T] {
	return HttpClientWraperResponse[T]{
		Err: err,
	}
}

func newHttpClientWraperResponseWithoutBody[T any](err error, statusCode int, status string) HttpClientWraperResponse[T] {
	return HttpClientWraperResponse[T]{
		StatusCode: statusCode,
		Status:     status,
		Err:        err,
	}
}

type HttpClientWrapper[T any] interface {
	Get(ctx context.Context, uri string, extraHeaders map[string]string) (HttpClientWraperResponse[T], error)
	Post(ctx context.Context, uri string, extraHeaders map[string]string, body interface{}) (HttpClientWraperResponse[T], error)
}

func httpClientDo[T any](ctx context.Context, httpClient *http.Client, httpClientTrace *HttpClientTraceWrapper, req *http.Request, extraHeaders map[string]string) (HttpClientWraperResponse[T], error) {
	for key, value := range extraHeaders {
		req.Header.Add(key, value)
	}

	// clientTrace := httpClientTrace.NewClientTrace()
	// req = req.WithContext(httptrace.WithClientTrace(req.Context(), clientTrace))
	resp, err := httpClient.Do(req)
	if err != nil {
		return newHttpClientWraperResponseWithoutResponse[T](
			err,
		), err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newHttpClientWraperResponseWithoutBody[T](
			fmt.Errorf("status code: %d", resp.StatusCode),
			resp.StatusCode,
			resp.Status,
		), fmt.Errorf("status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return newHttpClientWraperResponseWithoutBody[T](
			fmt.Errorf("status code: %d", resp.StatusCode),
			resp.StatusCode,
			resp.Status,
		), err
	}

	var response T
	if err := json.Unmarshal(body, &response); err != nil {
		return newHttpClientWraperResponseWithoutBody[T](
			fmt.Errorf("status code: %d", resp.StatusCode),
			resp.StatusCode,
			resp.Status,
		), err
	}

	return HttpClientWraperResponse[T]{
		Body:       response,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}, nil
}
