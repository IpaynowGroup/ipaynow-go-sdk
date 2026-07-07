package ipaynow

import "testing"

// Table-driven check of the funcode/version/path/signTypeField/excludeSignType
// metadata for every operation (spec §2).
func TestRequestMetadataMatchesSpec(t *testing.T) {
	cases := []struct {
		name            string
		meta            RequestMeta
		wantFuncode     string
		wantVersion     string
		wantPath        string
		wantSignField   string
		wantExcludeType bool
	}{
		{"jsapi", NewJsapiUnifiedOrderRequest(JsapiUnifiedOrderParams{}).RequestMeta, "WP001", "1.0.0", "/", "mhtSignType", false},
		{"aggregate", NewAggregateCodeUnifiedOrderRequest(AggregateCodeUnifiedOrderParams{}).RequestMeta, "WP001", "1.0.3", "/", "mhtSignType", false},
		{"miniprogram", NewMiniProgramUnifiedOrderRequest(MiniProgramUnifiedOrderParams{}).RequestMeta, "WP001", "1.0.0", "/", "mhtSignType", false},
		{"h5", NewH5UnifiedOrderRequest(H5UnifiedOrderParams{}).RequestMeta, "WP001", "1.0.0", "/", "mhtSignType", false},
		{"app", NewAppUnifiedOrderRequest(AppUnifiedOrderParams{}).RequestMeta, "WP001", "1.0.3", "/", "mhtSignType", false},
		{"orderquery", NewOrderQueryRequest(OrderQueryParams{}).RequestMeta, "MQ002", "1.0.0", "/", "mhtSignType", false},
		{"refund", NewRefundRequest(RefundParams{}).RequestMeta, "R001", "1.0.0", "/refund/refundOrder", "signType", true},
		{"refundquery", NewRefundQueryRequest(RefundQueryParams{}).RequestMeta, "Q001", "1.0.0", "/refund/refundQuery", "signType", true},
		{"reverse", NewReverseRequest(ReverseParams{}).RequestMeta, "R002", "1.0.0", "/refund/cancelOrder", "signType", true},
		{"reversequery", NewReverseQueryRequest(ReverseQueryParams{}).RequestMeta, "Q002", "1.0.0", "/refund/refundQuery", "signType", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.meta.Funcode != tc.wantFuncode {
				t.Errorf("funcode = %s, want %s", tc.meta.Funcode, tc.wantFuncode)
			}
			if tc.meta.Version != tc.wantVersion {
				t.Errorf("version = %s, want %s", tc.meta.Version, tc.wantVersion)
			}
			if tc.meta.Path != tc.wantPath {
				t.Errorf("path = %s, want %s", tc.meta.Path, tc.wantPath)
			}
			if tc.meta.SignTypeField != tc.wantSignField {
				t.Errorf("signTypeField = %s, want %s", tc.meta.SignTypeField, tc.wantSignField)
			}
			if tc.meta.ExcludeSignType != tc.wantExcludeType {
				t.Errorf("excludeSignType = %v, want %v", tc.meta.ExcludeSignType, tc.wantExcludeType)
			}
			if tc.meta.SignatureField != "mhtSignature" {
				t.Errorf("signatureField = %s, want mhtSignature", tc.meta.SignatureField)
			}
			if tc.meta.ResponseSignatureField != "signature" {
				t.Errorf("responseSignatureField = %s, want signature", tc.meta.ResponseSignatureField)
			}
		})
	}
}

func TestRequestVersionOverride(t *testing.T) {
	req := NewJsapiUnifiedOrderRequestWithVersion("1.0.2", JsapiUnifiedOrderParams{})
	if req.Version != "1.0.2" {
		t.Fatalf("Version = %s, want 1.0.2", req.Version)
	}
}

func TestOrderQueryRequestVersionOverrideFromParams(t *testing.T) {
	req := NewOrderQueryRequest(OrderQueryParams{Version: "2.0.0"})
	if req.Version != "2.0.0" {
		t.Fatalf("Version = %s, want 2.0.0 (overridden by params.Version)", req.Version)
	}
}

// --- Params defaults (spec §5) ---------------------------------------------

func TestJsapiUnifiedOrderParamsDefaults(t *testing.T) {
	m := JsapiUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 100}.ToMap()
	assertEqual(t, m, "deviceType", "0600")
	assertEqual(t, m, "mhtCurrencyType", "156")
	assertEqual(t, m, "mhtOrderTimeOut", "120")
	assertEqual(t, m, "mhtOrderType", "05")
	assertEqual(t, m, "mhtOrderAmt", "100")
	assertEqual(t, m, "mhtOrderNo", "O1")
}

func TestJsapiUnifiedOrderParamsCallerOverridesDefault(t *testing.T) {
	m := JsapiUnifiedOrderParams{MhtOrderNo: "O1", DeviceType: "9999", MhtOrderTimeOut: 60}.ToMap()
	assertEqual(t, m, "deviceType", "9999")
	assertEqual(t, m, "mhtOrderTimeOut", "60")
}

func TestAggregateCodeUnifiedOrderParamsDefaults(t *testing.T) {
	m := AggregateCodeUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1}.ToMap()
	assertEqual(t, m, "deviceType", "20")
}

func TestMiniProgramUnifiedOrderParamsDefaults(t *testing.T) {
	m := MiniProgramUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1}.ToMap()
	assertEqual(t, m, "deviceType", "14")
	assertEqual(t, m, "outputType", "1")
}

func TestMiniProgramUnifiedOrderParamsSupportsMhtStoreId(t *testing.T) {
	m := MiniProgramUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1, MhtStoreId: "S1"}.ToMap()
	assertEqual(t, m, "mhtStoreId", "S1")
}

func TestH5UnifiedOrderParamsDefaults(t *testing.T) {
	m := H5UnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1}.ToMap()
	assertEqual(t, m, "deviceType", "0601")
}

func TestH5UnifiedOrderParamsSupportsDocFields(t *testing.T) {
	m := H5UnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1, SpecifiedChannel: "wechat", PayAccNo: "18800000000", GuideId: "G1"}.ToMap()
	assertEqual(t, m, "specifiedChannel", "wechat")
	assertEqual(t, m, "payAccNo", "18800000000")
	assertEqual(t, m, "guideId", "G1")
}

func TestAppUnifiedOrderParamsDefaults(t *testing.T) {
	m := AppUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1}.ToMap()
	assertEqual(t, m, "deviceType", "01")
	assertEqual(t, m, "outputType", "0")
}

func TestAppUnifiedOrderParamsSupportsDocFields(t *testing.T) {
	m := AppUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1, FrontNotifyUrl: "https://example.com/front", MhtStoreId: "S1"}.ToMap()
	assertEqual(t, m, "frontNotifyUrl", "https://example.com/front")
	assertEqual(t, m, "mhtStoreId", "S1")
}

func TestParamsExtraParamsMergeAndOverride(t *testing.T) {
	m := JsapiUnifiedOrderParams{
		MhtOrderNo:  "O1",
		MhtOrderAmt: 1,
		ExtraParams: map[string]string{"deviceType": "override", "custom": "v"},
	}.ToMap()
	assertEqual(t, m, "deviceType", "override")
	assertEqual(t, m, "custom", "v")
}

func TestParamsOptionalPointerFields(t *testing.T) {
	m := MiniProgramUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1}.ToMap()
	if _, ok := m["discountAmt"]; ok {
		t.Fatal("discountAmt should be absent when nil")
	}
	amt := int64(50)
	m2 := MiniProgramUnifiedOrderParams{MhtOrderNo: "O1", MhtOrderAmt: 1, DiscountAmt: &amt}.ToMap()
	assertEqual(t, m2, "discountAmt", "50")
}

func assertEqual(t *testing.T, m map[string]string, key, want string) {
	t.Helper()
	if got := m[key]; got != want {
		t.Errorf("%s = %q, want %q", key, got, want)
	}
}
