package ipaynow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Transport sends a signed, form-encoded POST and returns the raw response
// body. It is an interface so tests can inject a fake without spinning up
// real sockets, and so the timeout-classification logic below stays out of
// Client.
type Transport interface {
	Post(rawURL string, params map[string]string) (string, error)
}

// connectTimeoutError tags a dial timeout so it can be distinguished from a
// later read timeout after it has been wrapped by net/http (*url.Error).
type connectTimeoutError struct{ err error }

func (e *connectTimeoutError) Error() string { return e.err.Error() }
func (e *connectTimeoutError) Timeout() bool { return true }
func (e *connectTimeoutError) Unwrap() error { return e.err }

// defaultTransport is the net/http-based Transport used in production.
type defaultTransport struct {
	client *http.Client
	logger *slog.Logger
}

// TransportOption customizes a defaultTransport at construction time.
type TransportOption func(*defaultTransport)

// withTransportLogger injects a *slog.Logger for ERROR-level HTTP
// failure/timeout logging (spec §8). Unexported: production wiring goes
// through Client's WithLogger option; NewDefaultTransport defaults to
// slog.Default() when used directly.
func withTransportLogger(logger *slog.Logger) TransportOption {
	return func(t *defaultTransport) { t.logger = logger }
}

// NewDefaultTransport builds a Transport whose DialContext timeout maps to
// CONNECT_TIMEOUT (E0003) and whose overall Client.Timeout maps to
// SOCKET_TIMEOUT (E0004) — unlike Java's OkHttp, Go lets us tell these two
// timeouts apart precisely instead of resorting to a message heuristic.
func NewDefaultTransport(connectTimeout, readTimeout time.Duration, opts ...TransportOption) Transport {
	dialer := &net.Dialer{Timeout: connectTimeout}
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					return nil, &connectTimeoutError{err: err}
				}
				return nil, err
			}
			return conn, nil
		},
	}
	t := &defaultTransport{client: &http.Client{Transport: tr, Timeout: readTimeout}, logger: slog.Default()}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *defaultTransport) Post(rawURL string, params map[string]string) (string, error) {
	funcode := params["funcode"]
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	resp, err := t.client.PostForm(rawURL, form)
	if err != nil {
		var cte *connectTimeoutError
		if errors.As(err, &cte) {
			t.logger.Error("现在支付请求建连超时", "funcode", funcode, "error", err)
			return "", &IpaynowError{Code: CodeConnectTimeout, Message: fmt.Sprintf("现在支付请求建连超时：%v", err), Cause: err}
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			t.logger.Error("现在支付请求读取超时", "funcode", funcode, "error", err)
			return "", &IpaynowError{Code: CodeSocketTimeout, Message: fmt.Sprintf("现在支付请求读取超时：%v", err), Cause: err}
		}
		t.logger.Error("现在支付请求异常", "funcode", funcode, "error", err)
		return "", &IpaynowError{Code: CodeHTTPException, Message: fmt.Sprintf("现在支付请求异常：%v", err), Cause: err}
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		var netErr net.Error
		if errors.As(readErr, &netErr) && netErr.Timeout() {
			t.logger.Error("现在支付响应读取超时", "funcode", funcode, "error", readErr)
			return "", &IpaynowError{Code: CodeSocketTimeout, Message: fmt.Sprintf("现在支付响应读取超时：%v", readErr), Cause: readErr}
		}
		t.logger.Error("读取现在支付响应体失败", "funcode", funcode, "error", readErr)
		return "", &IpaynowError{Code: CodeHTTPException, Message: fmt.Sprintf("读取现在支付响应体失败：%v", readErr), Cause: readErr}
	}
	text := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.logger.Error("现在支付HTTP请求失败", "funcode", funcode, "status", resp.StatusCode, "body", text)
		return "", &IpaynowError{Code: CodeHTTPException, Message: fmt.Sprintf("现在支付HTTP请求失败，状态码=%d，响应报文=%s", resp.StatusCode, text)}
	}
	return text, nil
}
