package ipaynow

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestExecuteAssemblesCommonFieldsAndReturnsTypedResponse(t *testing.T) {
	var capturedForm string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		capturedForm = r.Form.Encode()

		respFields := map[string]string{"responseCode": "A001", "tn": "T123", "nowPayOrderNo": "N1"}
		cfg, _ := NewConfig("myapp", "secret")
		sig := Sign(respFields, cfg, false, "signature", "mhtSignature")
		w.Write([]byte("responseCode=A001&tn=T123&nowPayOrderNo=N1&signature=" + sig))
	}))
	defer server.Close()

	cfg, err := NewConfig("myapp", "secret", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}
	client := NewClient(cfg)

	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})
	resp, err := Execute(client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatal("expected success response")
	}
	if resp.Tn() != "T123" {
		t.Fatalf("Tn() = %s, want T123", resp.Tn())
	}
	if resp.NowPayOrderNo() != "N1" {
		t.Fatalf("NowPayOrderNo() = %s, want N1", resp.NowPayOrderNo())
	}

	// Public/common fields must have been sent.
	if !strings.Contains(capturedForm, "funcode=WP001") {
		t.Errorf("form missing funcode: %s", capturedForm)
	}
	if !strings.Contains(capturedForm, "appId=myapp") {
		t.Errorf("form missing appId: %s", capturedForm)
	}
	if !strings.Contains(capturedForm, "mhtCharset=UTF-8") {
		t.Errorf("form missing mhtCharset: %s", capturedForm)
	}
	if !strings.Contains(capturedForm, "mhtSignType=MD5") {
		t.Errorf("form missing mhtSignType: %s", capturedForm)
	}
	if !strings.Contains(capturedForm, "mhtSignature=") {
		t.Errorf("form missing mhtSignature: %s", capturedForm)
	}
}

func TestExecuteReturnsSignErrorOnBadResponseSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("responseCode=A001&signature=deadbeef"))
	}))
	defer server.Close()

	cfg, _ := NewConfig("myapp", "secret", WithBaseURL(server.URL))
	client := NewClient(cfg)
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	_, err := Execute(client, req)
	assertSignError(t, err)
}

func TestExecuteReturnsHTTPExceptionOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	cfg, _ := NewConfig("myapp", "secret", WithBaseURL(server.URL))
	client := NewClient(cfg)
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	_, err := Execute(client, req)
	if err == nil {
		t.Fatal("expected an error")
	}
	ipErr, ok := err.(*IpaynowError)
	if !ok {
		t.Fatalf("expected *IpaynowError, got %T", err)
	}
	if ipErr.Code != CodeHTTPException {
		t.Fatalf("Code = %s, want %s", ipErr.Code, CodeHTTPException)
	}
	if !strings.Contains(ipErr.Message, "500") || !strings.Contains(ipErr.Message, "boom") {
		t.Fatalf("message should contain status code and response body: %s", ipErr.Message)
	}
}

func TestExecuteReturnsSocketTimeoutOnSlowResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		w.Write([]byte("responseCode=A001"))
	}))
	defer server.Close()

	cfg, _ := NewConfig("myapp", "secret",
		WithBaseURL(server.URL),
		WithConnectTimeout(2*time.Second),
		WithReadTimeout(20*time.Millisecond),
	)
	client := NewClient(cfg)
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	_, err := Execute(client, req)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	ipErr, ok := err.(*IpaynowError)
	if !ok {
		t.Fatalf("expected *IpaynowError, got %T", err)
	}
	if ipErr.Code != CodeSocketTimeout {
		t.Fatalf("Code = %s, want %s", ipErr.Code, CodeSocketTimeout)
	}
}

func TestExecuteReturnsConnectTimeoutOnUnroutableAddress(t *testing.T) {
	// 10.255.255.1 is a non-routable address commonly used to force a dial
	// timeout without depending on external network flakiness.
	cfg, _ := NewConfig("myapp", "secret",
		WithBaseURL("http://10.255.255.1"),
		WithConnectTimeout(50*time.Millisecond),
		WithReadTimeout(5*time.Second),
	)
	client := NewClient(cfg)
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	_, err := Execute(client, req)
	if err == nil {
		t.Skip("dial did not time out in this environment; skipping")
	}
	ipErr, ok := err.(*IpaynowError)
	if !ok {
		t.Fatalf("expected *IpaynowError, got %T", err)
	}
	if ipErr.Code != CodeConnectTimeout {
		t.Fatalf("Code = %s, want %s (got message: %s)", ipErr.Code, CodeConnectTimeout, ipErr.Message)
	}
}

func TestBuildFormFieldsDoesNotSendRequest(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	cfg, _ := NewConfig("myapp", "secret", WithBaseURL(server.URL))
	client := NewClient(cfg)
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	fields := BuildFormFields(client, req)
	if called {
		t.Fatal("BuildFormFields must not perform an HTTP request")
	}
	if fields["mhtOrderNo"] != "O1" {
		t.Fatalf("fields missing mhtOrderNo: %#v", fields)
	}
	if fields["mhtSignature"] == "" {
		t.Fatal("fields missing mhtSignature")
	}
}

func TestExecuteUsesRSASignTypeFromConfig(t *testing.T) {
	cfg := rsaConfig(t)
	cfg.BaseURL = "https://pay.test"
	transport := &captureTransport{
		body: func(_ map[string]string) string {
			respFields := map[string]string{"responseCode": "R000", "signType": "RSA"}
			sig := Sign(respFields, cfg, true, "signature", "mhtSignature")
			return "responseCode=R000&signType=RSA&signature=" + url.QueryEscape(sig)
		},
	}
	client := NewClientWithTransport(cfg, transport)
	req := NewRefundRequest(RefundParams{MhtOrderNo: "O1", MhtRefundNo: "R1", Amount: 1})

	resp, err := Execute(client, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if transport.params["signType"] != "RSA" || !resp.IsSuccess() {
		t.Fatalf("unexpected RSA flow: params=%v resp=%#v", transport.params, resp)
	}
}

type captureTransport struct {
	params map[string]string
	body   func(map[string]string) string
}

func (c *captureTransport) Post(_ string, params map[string]string) (string, error) {
	c.params = params
	return c.body(params), nil
}
