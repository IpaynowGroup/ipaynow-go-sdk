package ipaynow

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"sort"
	"strings"
)

// Sign implements the canonical iPayNow signature algorithm (spec §3):
//  1. drop null/empty values
//  2. drop signature fields (mhtSignature, signature, and any caller-supplied
//     excludeFields such as the request's own signature/signType field names)
//  3. if excludeSignType, additionally drop "signType" (mhtSignType always
//     participates - it is a distinct field name and is never excluded here)
//  4. sort remaining keys ascending
//  5. join as k1=v1&k2=v2&... (values are not URL-encoded)
//  6. append "&" + md5(appKey)
//  7. md5 the whole string, lowercase hex
func Sign(params map[string]string, config Config, excludeSignType bool, excludeFields ...string) string {
	canonical := canonical(params, excludeSignType, excludeFields...)
	if config.SignType == SignTypeRSA {
		signature, err := rsa.SignPKCS1v15(rand.Reader, config.MchPrivateKeyObject, crypto.SHA256, sha256Bytes(canonical))
		if err != nil {
			panic(err)
		}
		return base64Encode(signature)
	}
	return md5Hex(canonical + "&" + md5Hex(config.AppKey))
}

// Verify checks a response/notification signature. It fails closed: a
// missing or empty signature field always returns false (spec §4 red line 1),
// and the comparison is constant-time (red line 2).
func Verify(fields map[string]string, config Config, signatureField string, excludeSignType bool) bool {
	actual, ok := fields[signatureField]
	if !ok || actual == "" {
		return false
	}
	if actualSignType, ok := fields["signType"]; ok && actualSignType != "" && !strings.EqualFold(actualSignType, string(config.SignType)) {
		return false
	}
	canonical := canonical(fields, excludeSignType, signatureField, "mhtSignature")
	if config.SignType == SignTypeRSA {
		signature, err := base64Decode(actual)
		if err != nil {
			return false
		}
		return rsa.VerifyPKCS1v15(config.IpaynowPublicKeyObject, crypto.SHA256, sha256Bytes(canonical), signature) == nil
	}
	expected := Sign(fields, config, excludeSignType, signatureField, "mhtSignature")
	a := []byte(expected)
	b := []byte(strings.ToLower(actual))
	return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1
}

// clean returns a copy of params with null/empty-string values removed.
func clean(params map[string]string) map[string]string {
	out := make(map[string]string, len(params))
	for k, v := range params {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func md5Hex(value string) string {
	sum := md5.Sum([]byte(value))
	return hex.EncodeToString(sum[:])
}

func canonical(params map[string]string, excludeSignType bool, excludeFields ...string) string {
	cleaned := clean(params)
	if excludeSignType {
		delete(cleaned, "signType")
	}
	delete(cleaned, "mhtSignature")
	delete(cleaned, "signature")
	for _, field := range excludeFields {
		delete(cleaned, field)
	}

	keys := make([]string, 0, len(cleaned))
	for k := range cleaned {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out strings.Builder
	for i, k := range keys {
		if i > 0 {
			out.WriteByte('&')
		}
		out.WriteString(k)
		out.WriteByte('=')
		out.WriteString(cleaned[k])
	}
	return out.String()
}

func sha256Bytes(value string) []byte {
	sum := crypto.SHA256.New()
	sum.Write([]byte(value))
	return sum.Sum(nil)
}

func base64Encode(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

func base64Decode(value string) ([]byte, error) {
	return decodeBase64(value, "signature")
}
