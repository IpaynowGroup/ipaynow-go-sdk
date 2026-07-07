package ipaynow

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

const defaultBaseURL = "https://pay.ipaynow.cn"

type SignType string

const (
	SignTypeMD5 SignType = "MD5"
	SignTypeRSA SignType = "RSA"
)

// Config holds merchant credentials and transport tuning. Build it with
// NewConfig, which validates inputs and fills in the documented defaults.
type Config struct {
	AppID                  string
	AppKey                 string
	MchPrivateKey          string
	IpaynowPublicKey       string
	MchPrivateKeyObject    *rsa.PrivateKey
	IpaynowPublicKeyObject *rsa.PublicKey
	SignType               SignType
	BaseURL                string
	ConnectTimeout         time.Duration
	ReadTimeout            time.Duration
}

// Option customizes a Config before validation in NewConfig.
type Option func(*Config)

// WithBaseURL overrides the default gateway base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) { c.BaseURL = baseURL }
}

// WithConnectTimeout overrides the default TCP connect timeout (5s).
func WithConnectTimeout(d time.Duration) Option {
	return func(c *Config) { c.ConnectTimeout = d }
}

// WithReadTimeout overrides the default response read timeout (15s).
func WithReadTimeout(d time.Duration) Option {
	return func(c *Config) { c.ReadTimeout = d }
}

func WithSignType(signType SignType) Option {
	return func(c *Config) { c.SignType = signType }
}

func WithMerchantPrivateKey(privateKey string) Option {
	return func(c *Config) { c.MchPrivateKey = privateKey }
}

func WithIpaynowPublicKey(publicKey string) Option {
	return func(c *Config) { c.IpaynowPublicKey = publicKey }
}

// NewConfig builds a validated Config. appId, appKey and baseUrl must be
// non-blank and both timeouts must be positive; violations return a plain
// error (mirroring the Java SDK's IllegalArgumentException, which is
// distinct from the IpaynowError codes reserved for runtime SDK failures).
func NewConfig(appID, appKey string, opts ...Option) (Config, error) {
	cfg := Config{
		AppID:          appID,
		AppKey:         appKey,
		SignType:       SignTypeMD5,
		BaseURL:        defaultBaseURL,
		ConnectTimeout: 5000 * time.Millisecond,
		ReadTimeout:    15000 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if strings.TrimSpace(cfg.AppID) == "" {
		return Config{}, fmt.Errorf("ipaynow: appId 不能为空")
	}
	if cfg.SignType == SignTypeMD5 {
		if strings.TrimSpace(cfg.AppKey) == "" {
			return Config{}, fmt.Errorf("ipaynow: appKey 不能为空")
		}
	} else if cfg.SignType == SignTypeRSA {
		if strings.TrimSpace(cfg.MchPrivateKey) == "" {
			return Config{}, fmt.Errorf("ipaynow: mchPrivateKey 不能为空")
		}
		if strings.TrimSpace(cfg.IpaynowPublicKey) == "" {
			return Config{}, fmt.Errorf("ipaynow: ipaynowPublicKey 不能为空")
		}
		privateKey, err := parsePrivateKey(cfg.MchPrivateKey)
		if err != nil {
			return Config{}, err
		}
		publicKey, err := parsePublicKey(cfg.IpaynowPublicKey)
		if err != nil {
			return Config{}, err
		}
		cfg.MchPrivateKeyObject = privateKey
		cfg.IpaynowPublicKeyObject = publicKey
	} else {
		return Config{}, fmt.Errorf("ipaynow: 不支持的签名类型: %s", cfg.SignType)
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return Config{}, fmt.Errorf("ipaynow: baseUrl 不能为空")
	}
	if cfg.ConnectTimeout <= 0 || cfg.ReadTimeout <= 0 {
		return Config{}, fmt.Errorf("ipaynow: 超时时间必须大于 0：connectTimeout=%s，readTimeout=%s", cfg.ConnectTimeout, cfg.ReadTimeout)
	}
	return cfg, nil
}

// endpoint joins the configured base URL with a request path.
func (c Config) endpoint(path string) string {
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(c.BaseURL, "/") + path
}

func parsePrivateKey(value string) (*rsa.PrivateKey, error) {
	der, err := decodeBase64(value, "mchPrivateKey")
	if err != nil {
		return nil, err
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("ipaynow: mchPrivateKey 非法，需传裸 base64 PKCS8 私钥: %w", err)
	}
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("ipaynow: mchPrivateKey 非法，需传裸 base64 PKCS8 私钥")
	}
	return rsaKey, nil
}

func parsePublicKey(value string) (*rsa.PublicKey, error) {
	der, err := decodeBase64(value, "ipaynowPublicKey")
	if err != nil {
		return nil, err
	}
	publicKey, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("ipaynow: ipaynowPublicKey 非法，需传裸 base64 X509 公钥: %w", err)
	}
	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("ipaynow: ipaynowPublicKey 非法，需传裸 base64 X509 公钥")
	}
	return rsaKey, nil
}

func decodeBase64(value, name string) ([]byte, error) {
	if strings.Contains(value, "-----BEGIN") || strings.Contains(value, "-----END") {
		return nil, fmt.Errorf("ipaynow: %s 仅支持裸 base64，不支持 PEM", name)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(value), ""))
	if err != nil {
		return nil, fmt.Errorf("ipaynow: %s 非法，需传裸 base64: %w", name, err)
	}
	return decoded, nil
}
