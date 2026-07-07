package ipaynow

import (
	"fmt"
	"log/slog"
)

// Client executes signed requests against the iPayNow gateway.
type Client struct {
	config    Config
	transport Transport
	logger    *slog.Logger
}

// ClientOption customizes a Client at construction time.
type ClientOption func(*Client)

// WithLogger injects a *slog.Logger for request/response and error logging
// (spec §8). If not supplied, slog.Default() is used.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) { c.logger = logger }
}

// NewClient builds a Client using the production net/http Transport, tuned
// with config's connect/read timeouts.
func NewClient(config Config, opts ...ClientOption) *Client {
	c := &Client{config: config, logger: slog.Default()}
	for _, opt := range opts {
		opt(c)
	}
	c.transport = NewDefaultTransport(config.ConnectTimeout, config.ReadTimeout, withTransportLogger(c.logger))
	return c
}

// NewClientWithTransport builds a Client with an injected Transport, mainly
// for tests.
func NewClientWithTransport(config Config, transport Transport, opts ...ClientOption) *Client {
	c := &Client{config: config, transport: transport, logger: slog.Default()}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// signedFields assembles the common fields (funcode/version/appId/
// mhtCharset/signType field/signature) on top of the operation params and
// returns the cleaned (empty-value-stripped) result.
func (c *Client) signedFields(meta RequestMeta) map[string]string {
	params := make(map[string]string, len(meta.Params)+5)
	for k, v := range meta.Params {
		params[k] = v
	}
	params["funcode"] = meta.Funcode
	params["version"] = meta.Version
	params["appId"] = c.config.AppID
	params["mhtCharset"] = meta.MhtCharset
	params[meta.SignTypeField] = string(c.config.SignType)
	params[meta.SignatureField] = Sign(params, c.config, meta.ExcludeSignType, meta.SignatureField, meta.ResponseSignatureField)
	return clean(params)
}

// BuildFormFields returns the fully signed form fields for req without
// sending a request — useful for front-end redirect/jump forms.
func BuildFormFields[T any](c *Client, req *Request[T]) map[string]string {
	return c.signedFields(req.RequestMeta)
}

// Execute sends req, parses the response, verifies its signature, and
// returns the typed Response. A missing/invalid response signature returns
// an *IpaynowError with CodeSignError; it is never silently ignored.
func Execute[T any](c *Client, req *Request[T]) (*T, error) {
	fields := c.signedFields(req.RequestMeta)
	// 报文可能含用户/订单敏感信息，只在 DEBUG 级别输出，且签名值脱敏。
	c.logger.Debug("现在支付请求报文", "funcode", req.Funcode, "fields", redactSensitive(fields))

	body, err := c.transport.Post(c.config.endpoint(req.Path), fields)
	if err != nil {
		return nil, err
	}

	respFields := parseResponse(body)
	c.logger.Debug("现在支付响应报文", "funcode", req.Funcode, "fields", redactSensitive(respFields))
	if !Verify(respFields, c.config, req.ResponseSignatureField, req.ExcludeSignType) {
		c.logger.Warn("现在支付验签失败", "funcode", req.Funcode)
		return nil, &IpaynowError{Code: CodeSignError, Message: fmt.Sprintf("现在支付响应验签失败，功能码=%s", req.Funcode)}
	}

	resp := req.newResponse(respFields, body)
	return &resp, nil
}
