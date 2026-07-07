package ipaynow

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRedactSensitiveMasksSignatureFields(t *testing.T) {
	fields := map[string]string{
		"mhtOrderNo":   "O1",
		"mhtSignature": "abc123",
		"signature":    "def456",
		"mhtOrderAmt":  "100",
	}
	redacted := redactSensitive(fields)

	if redacted["mhtSignature"] != "***" {
		t.Fatalf("mhtSignature = %s, want ***", redacted["mhtSignature"])
	}
	if redacted["signature"] != "***" {
		t.Fatalf("signature = %s, want ***", redacted["signature"])
	}
	if redacted["mhtOrderNo"] != "O1" {
		t.Fatalf("mhtOrderNo = %s, want unchanged O1", redacted["mhtOrderNo"])
	}
	if redacted["mhtOrderAmt"] != "100" {
		t.Fatalf("mhtOrderAmt = %s, want unchanged 100", redacted["mhtOrderAmt"])
	}
	// The original map must not be mutated.
	if fields["mhtSignature"] != "abc123" {
		t.Fatalf("original fields mutated: mhtSignature = %s", fields["mhtSignature"])
	}
}

func TestExecuteDebugLogDoesNotContainAppKey(t *testing.T) {
	cfg, err := NewConfig("myapp", "topsecretkey")
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respFields := map[string]string{"responseCode": "A001", "tn": "T123", "nowPayOrderNo": "N1"}
		sig := Sign(respFields, cfg, false, "signature", "mhtSignature")
		w.Write([]byte("responseCode=A001&tn=T123&nowPayOrderNo=N1&signature=" + sig))
	}))
	defer server.Close()
	cfg.BaseURL = server.URL

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client := NewClient(cfg, WithLogger(logger))

	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})
	if _, err := Execute(client, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected DEBUG log output, got none")
	}
	if strings.Contains(output, "topsecretkey") {
		t.Fatalf("log output must never contain appKey: %s", output)
	}
	if !strings.Contains(output, "***") {
		t.Fatalf("log output must mask mhtSignature/signature with ***: %s", output)
	}
	if !strings.Contains(output, "现在支付请求报文") || !strings.Contains(output, "现在支付响应报文") {
		t.Fatalf("expected request and response DEBUG log lines: %s", output)
	}
}

func TestExecuteWarnLogOnVerifyFailureDoesNotContainAppKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("responseCode=A001&signature=deadbeef"))
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, _ := NewConfig("myapp", "topsecretkey", WithBaseURL(server.URL))
	client := NewClient(cfg, WithLogger(logger))
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	if _, err := Execute(client, req); err == nil {
		t.Fatal("expected sign error")
	}

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Fatalf("expected a WARN log line on verify failure: %s", output)
	}
	if strings.Contains(output, "topsecretkey") {
		t.Fatalf("log output must never contain appKey: %s", output)
	}
}

func TestTransportErrorLogOnNon2xxDoesNotContainAppKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg, _ := NewConfig("myapp", "topsecretkey", WithBaseURL(server.URL))
	client := NewClient(cfg, WithLogger(logger))
	req := NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100})

	if _, err := Execute(client, req); err == nil {
		t.Fatal("expected an HTTP error")
	}

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Fatalf("expected an ERROR log line on non-2xx response: %s", output)
	}
	if !strings.Contains(output, "500") || !strings.Contains(output, "boom") {
		t.Fatalf("ERROR log should include status code and response body: %s", output)
	}
	if !strings.Contains(output, "WP001") {
		t.Fatalf("ERROR log should include funcode: %s", output)
	}
	if strings.Contains(output, "topsecretkey") {
		t.Fatalf("log output must never contain appKey: %s", output)
	}
}
