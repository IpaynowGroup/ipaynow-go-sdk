package ipaynow

import (
	"net/url"
	"strings"
)

// Response is the base type embedded by every operation-specific response.
// It gives generic field access (Get/Fields/RawBody/Code/Message) plus the
// default isSuccess rule (responseCode == "A001"), which operations that
// need a different rule override by shadowing IsSuccess on their own type.
type Response struct {
	fields  map[string]string
	rawBody string
}

// load populates the response from parsed fields. Unexported: only the
// Client and NotifyParser hydrate responses, callers never call it directly.
func (r *Response) load(fields map[string]string, rawBody string) {
	cp := make(map[string]string, len(fields))
	for k, v := range fields {
		cp[k] = v
	}
	r.fields = cp
	r.rawBody = rawBody
}

// Get returns a single response field, or "" if absent.
func (r *Response) Get(name string) string { return r.fields[name] }

// Fields returns a copy of all response fields.
func (r *Response) Fields() map[string]string {
	cp := make(map[string]string, len(r.fields))
	for k, v := range r.fields {
		cp[k] = v
	}
	return cp
}

// RawBody returns the raw (undecoded) response body string.
func (r *Response) RawBody() string { return r.rawBody }

// Code returns the responseCode field.
func (r *Response) Code() string { return r.Get("responseCode") }

// Message returns the responseMsg field.
func (r *Response) Message() string { return r.Get("responseMsg") }

// IsSuccess implements the default synchronous-response success rule
// (spec §6): responseCode == "A001". Refund-family and notify responses
// override this with their own rule.
func (r *Response) IsSuccess() bool { return r.Code() == "A001" }

// UnifiedOrderResponse is embedded by the five WP001 unified-order
// responses; it adds the shared Tn (falls back to the legacy uppercase
// "TN" field) and NowPayOrderNo accessors from spec §6.
type UnifiedOrderResponse struct {
	Response
}

// Tn returns the payment credential ("tn"), falling back to the legacy
// uppercase "TN" field for compatibility.
func (r *UnifiedOrderResponse) Tn() string {
	if v := r.Get("tn"); v != "" {
		return v
	}
	return r.Get("TN")
}

// NowPayOrderNo returns the iPayNow transaction serial number.
func (r *UnifiedOrderResponse) NowPayOrderNo() string { return r.Get("nowPayOrderNo") }

// JsapiUnifiedOrderResponse is the WP001 (JSAPI) response.
type JsapiUnifiedOrderResponse struct{ UnifiedOrderResponse }

// AggregateCodeUnifiedOrderResponse is the WP001 (aggregate code) response.
type AggregateCodeUnifiedOrderResponse struct{ UnifiedOrderResponse }

// MiniProgramUnifiedOrderResponse is the WP001 (mini program) response.
type MiniProgramUnifiedOrderResponse struct{ UnifiedOrderResponse }

// H5UnifiedOrderResponse is the WP001 (H5) response.
type H5UnifiedOrderResponse struct{ UnifiedOrderResponse }

// AppUnifiedOrderResponse is the WP001 (APP) response.
type AppUnifiedOrderResponse struct{ UnifiedOrderResponse }

// OrderQueryResponse is the MQ002 response.
type OrderQueryResponse struct {
	Response
}

// TransStatus returns the transaction status field.
func (r *OrderQueryResponse) TransStatus() string { return r.Get("transStatus") }

// NowPayOrderNo returns the iPayNow transaction serial number.
func (r *OrderQueryResponse) NowPayOrderNo() string { return r.Get("nowPayOrderNo") }

// PayTime returns the payment completion time.
func (r *OrderQueryResponse) PayTime() string { return r.Get("payTime") }

// ChannelOrderNo returns the channel-side order number.
func (r *OrderQueryResponse) ChannelOrderNo() string { return r.Get("channelOrderNo") }

// RefundResponse is the R001 response. isSuccess uses the refund-family
// rule: responseCode == "R000" (spec §6).
type RefundResponse struct{ Response }

func (r *RefundResponse) IsSuccess() bool { return r.Code() == "R000" }

// RefundQueryResponse is the Q001 response.
type RefundQueryResponse struct{ Response }

func (r *RefundQueryResponse) IsSuccess() bool { return r.Code() == "R000" }

// ReverseResponse is the R002 response.
type ReverseResponse struct{ Response }

func (r *ReverseResponse) IsSuccess() bool { return r.Code() == "R000" }

// ReverseQueryResponse is the Q002 response.
type ReverseQueryResponse struct{ Response }

func (r *ReverseQueryResponse) IsSuccess() bool { return r.Code() == "R000" }

// PaymentNotifyResponse is the N001 payment callback. isSuccess uses
// transStatus == "A001" (spec §6), not responseCode.
type PaymentNotifyResponse struct{ Response }

func (r *PaymentNotifyResponse) IsSuccess() bool { return r.Get("transStatus") == "A001" }

// RefundNotifyResponse is the RN001 refund callback. isSuccess uses
// tradeStatus == "A001" (spec §6), not responseCode.
type RefundNotifyResponse struct{ Response }

func (r *RefundNotifyResponse) IsSuccess() bool { return r.Get("tradeStatus") == "A001" }

// parseResponse decodes a "k=v&k2=v2" form-encoded response body into a
// field map, URL-decoding both keys and values.
func parseResponse(text string) map[string]string {
	trimmed := strings.TrimSpace(text)
	out := make(map[string]string)
	if trimmed == "" {
		return out
	}
	for _, part := range strings.Split(trimmed, "&") {
		kv := strings.SplitN(part, "=", 2)
		key, _ := url.QueryUnescape(kv[0])
		value := ""
		if len(kv) > 1 {
			value, _ = url.QueryUnescape(kv[1])
		}
		out[key] = value
	}
	return out
}
