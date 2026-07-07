package ipaynow

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func md5Config(t *testing.T) Config {
	t.Helper()
	cfg, err := NewConfig("app", "secret")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	return cfg
}

func rsaConfig(t *testing.T) Config {
	t.Helper()
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
	return cfg
}

// Known vector from spec §10.1.
func TestSignKnownVector(t *testing.T) {
	params := map[string]string{"b": "2", "a": "1", "empty": "", "signType": "MD5", "mhtSignature": "bad"}
	got := Sign(params, md5Config(t), true)
	want := "c319c9529d781b179bff6bfdc3f82202"
	if got != want {
		t.Fatalf("Sign() = %s, want %s", got, want)
	}
}

func TestSignExcludesSignTypeOnlyWhenRequested(t *testing.T) {
	base := map[string]string{"a": "1", "signType": "MD5"}

	config := md5Config(t)
	excluded := Sign(base, config, true)
	included := Sign(base, config, false)
	if excluded == included {
		t.Fatal("expected different signatures when signType is included vs excluded")
	}

	// Sanity: excluding signType from `base` should match signing the map
	// without signType at all (with excludeSignType=false).
	withoutSignType := map[string]string{"a": "1"}
	if excluded != Sign(withoutSignType, config, false) {
		t.Fatal("excludeSignType=true should behave like signType being absent")
	}
}

func TestSignMhtSignTypeAlwaysParticipates(t *testing.T) {
	base := map[string]string{"a": "1", "mhtSignType": "MD5"}
	config := md5Config(t)
	withExclude := Sign(base, config, true)
	withoutField := Sign(map[string]string{"a": "1"}, config, true)
	if withExclude == withoutField {
		t.Fatal("mhtSignType must always participate in the signature, even when excludeSignType=true")
	}
}

func TestSignDropsEmptyAndSignatureFields(t *testing.T) {
	config := md5Config(t)
	a := Sign(map[string]string{"a": "1", "empty": "", "mhtSignature": "x", "signature": "y"}, config, false)
	b := Sign(map[string]string{"a": "1"}, config, false)
	if a != b {
		t.Fatalf("expected empty/signature fields to be dropped from signing input: %s != %s", a, b)
	}
}

func TestVerifyMissingSignatureFails(t *testing.T) {
	config := md5Config(t)
	if Verify(map[string]string{"a": "1"}, config, "signature", false) {
		t.Fatal("verify must fail closed when the signature field is missing")
	}
	if Verify(map[string]string{"a": "1", "signature": ""}, config, "signature", false) {
		t.Fatal("verify must fail closed when the signature field is empty")
	}
}

func TestVerifyForwardResponseUsesSignTypeIncluded(t *testing.T) {
	fields := map[string]string{"responseCode": "A001", "signType": "MD5"}
	config := md5Config(t)
	sig := Sign(fields, config, false, "signature", "mhtSignature")
	fields["signature"] = sig

	if !Verify(fields, config, "signature", false) {
		t.Fatal("forward-transaction response should verify with excludeSignType=false")
	}
	if Verify(fields, config, "signature", true) {
		t.Fatal("forward-transaction response should NOT verify with excludeSignType=true")
	}
}

func TestVerifyRefundResponseUsesSignTypeExcluded(t *testing.T) {
	fields := map[string]string{"responseCode": "R000", "signType": "MD5"}
	config := md5Config(t)
	sig := Sign(fields, config, true, "signature", "mhtSignature")
	fields["signature"] = sig

	if !Verify(fields, config, "signature", true) {
		t.Fatal("refund-family response should verify with excludeSignType=true")
	}
	if Verify(fields, config, "signature", false) {
		t.Fatal("refund-family response should NOT verify with excludeSignType=false")
	}
}

func TestVerifyIsCaseInsensitiveOnSignatureValue(t *testing.T) {
	fields := map[string]string{"responseCode": "A001"}
	config := md5Config(t)
	sig := Sign(fields, config, false, "signature", "mhtSignature")
	fields["signature"] = upper(sig)
	if !Verify(fields, config, "signature", false) {
		t.Fatal("verify should accept an uppercase signature value")
	}
}

func TestVerifyRejectsSignTypeMismatch(t *testing.T) {
	config := md5Config(t)
	fields := map[string]string{"responseCode": "A001", "signType": "RSA"}
	fields["signature"] = Sign(fields, config, false, "signature", "mhtSignature")
	if Verify(fields, config, "signature", false) {
		t.Fatal("verify should fail when signType mismatches config")
	}
}

func TestSignAndVerifyRSA(t *testing.T) {
	config := rsaConfig(t)
	fields := map[string]string{"responseCode": "A001", "signType": "RSA"}
	fields["signature"] = Sign(fields, config, false, "signature", "mhtSignature")
	if !Verify(fields, config, "signature", false) {
		t.Fatal("RSA sign/verify should succeed")
	}
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}
