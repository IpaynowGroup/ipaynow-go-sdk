package ipaynow

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
	"time"
)

func TestNewConfigDefaults(t *testing.T) {
	cfg, err := NewConfig("app", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != defaultBaseURL {
		t.Fatalf("BaseURL = %s, want %s", cfg.BaseURL, defaultBaseURL)
	}
	if cfg.ConnectTimeout != 5000*time.Millisecond {
		t.Fatalf("ConnectTimeout = %s, want 5000ms", cfg.ConnectTimeout)
	}
	if cfg.ReadTimeout != 15000*time.Millisecond {
		t.Fatalf("ReadTimeout = %s, want 15000ms", cfg.ReadTimeout)
	}
}

func TestNewConfigRejectsBlankAppID(t *testing.T) {
	if _, err := NewConfig("  ", "key"); err == nil {
		t.Fatal("expected error for blank appId")
	}
}

func TestNewConfigRejectsBlankAppKey(t *testing.T) {
	if _, err := NewConfig("app", " "); err == nil {
		t.Fatal("expected error for blank appKey")
	}
}

func TestNewConfigRejectsBlankBaseURL(t *testing.T) {
	if _, err := NewConfig("app", "key", WithBaseURL(" ")); err == nil {
		t.Fatal("expected error for blank baseUrl")
	}
}

func TestNewConfigRejectsNonPositiveTimeouts(t *testing.T) {
	if _, err := NewConfig("app", "key", WithConnectTimeout(0)); err == nil {
		t.Fatal("expected error for zero connect timeout")
	}
	if _, err := NewConfig("app", "key", WithReadTimeout(-1*time.Millisecond)); err == nil {
		t.Fatal("expected error for negative read timeout")
	}
}

func TestConfigEndpointJoinsBaseURLAndPath(t *testing.T) {
	cfg, err := NewConfig("app", "key", WithBaseURL("https://pay.ipaynow.cn/"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cfg.endpoint("/refund/refundOrder"); got != "https://pay.ipaynow.cn/refund/refundOrder" {
		t.Fatalf("endpoint = %s", got)
	}
	if got := cfg.endpoint("/"); got != "https://pay.ipaynow.cn/" {
		t.Fatalf("endpoint = %s", got)
	}
}

func TestNewConfigRSAParsesBase64AndRejectsPEM(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey() error = %v", err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}

	cfg, err := NewConfig(
		"app",
		"",
		WithSignType(SignTypeRSA),
		WithMerchantPrivateKey(base64.StdEncoding.EncodeToString(privateDER)),
		WithIpaynowPublicKey(base64.StdEncoding.EncodeToString(publicDER)),
	)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.SignType != SignTypeRSA || cfg.AppKey != "" {
		t.Fatalf("unexpected RSA config: %#v", cfg)
	}
	if cfg.MchPrivateKeyObject == nil || cfg.IpaynowPublicKeyObject == nil {
		t.Fatal("RSA key objects should be parsed eagerly")
	}

	if _, err := NewConfig(
		"app",
		"",
		WithSignType(SignTypeRSA),
		WithMerchantPrivateKey("-----BEGIN PRIVATE KEY-----bad"),
		WithIpaynowPublicKey(base64.StdEncoding.EncodeToString(publicDER)),
	); err == nil {
		t.Fatal("expected PEM merchant private key to be rejected")
	}
}
