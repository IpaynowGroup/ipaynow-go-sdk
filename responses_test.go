package ipaynow

import "testing"

// isSuccess rules per spec §6.
func TestIsSuccessRules(t *testing.T) {
	var jsapi JsapiUnifiedOrderResponse
	jsapi.load(map[string]string{"responseCode": "A001"}, "")
	if !jsapi.IsSuccess() {
		t.Fatal("unified order response should succeed on responseCode=A001")
	}
	jsapi.load(map[string]string{"responseCode": "B001"}, "")
	if jsapi.IsSuccess() {
		t.Fatal("unified order response should fail on non-A001 responseCode")
	}

	var oq OrderQueryResponse
	oq.load(map[string]string{"responseCode": "A001"}, "")
	if !oq.IsSuccess() {
		t.Fatal("order query response should succeed on responseCode=A001")
	}

	var refund RefundResponse
	refund.load(map[string]string{"responseCode": "R000"}, "")
	if !refund.IsSuccess() {
		t.Fatal("refund response should succeed on responseCode=R000")
	}
	refund.load(map[string]string{"responseCode": "A001"}, "")
	if refund.IsSuccess() {
		t.Fatal("refund response should NOT succeed on responseCode=A001")
	}

	var refundQuery RefundQueryResponse
	refundQuery.load(map[string]string{"responseCode": "R000"}, "")
	if !refundQuery.IsSuccess() {
		t.Fatal("refund query response should succeed on responseCode=R000")
	}

	var reverse ReverseResponse
	reverse.load(map[string]string{"responseCode": "R000"}, "")
	if !reverse.IsSuccess() {
		t.Fatal("reverse response should succeed on responseCode=R000")
	}

	var reverseQuery ReverseQueryResponse
	reverseQuery.load(map[string]string{"responseCode": "R000"}, "")
	if !reverseQuery.IsSuccess() {
		t.Fatal("reverse query response should succeed on responseCode=R000")
	}

	var paymentNotify PaymentNotifyResponse
	paymentNotify.load(map[string]string{"responseCode": "A001", "transStatus": "A001"}, "")
	if !paymentNotify.IsSuccess() {
		t.Fatal("payment notify should succeed on transStatus=A001")
	}
	paymentNotify.load(map[string]string{"responseCode": "A001", "transStatus": "B001"}, "")
	if paymentNotify.IsSuccess() {
		t.Fatal("payment notify should NOT succeed when transStatus != A001, even if responseCode=A001")
	}
	paymentNotify.load(map[string]string{"tradeStatus": "A001"}, "")
	if !paymentNotify.IsSuccess() {
		t.Fatal("payment notify should succeed on tradeStatus=A001 when transStatus is absent")
	}
	paymentNotify.load(map[string]string{"tradeStatus": "B001"}, "")
	if paymentNotify.IsSuccess() {
		t.Fatal("payment notify should NOT succeed when tradeStatus != A001")
	}
	paymentNotify.load(map[string]string{"transStatus": "B001", "tradeStatus": "A001"}, "")
	if paymentNotify.IsSuccess() {
		t.Fatal("payment notify should prefer transStatus over tradeStatus")
	}

	var refundNotify RefundNotifyResponse
	refundNotify.load(map[string]string{"tradeStatus": "A001"}, "")
	if !refundNotify.IsSuccess() {
		t.Fatal("refund notify should succeed on tradeStatus=A001")
	}
	refundNotify.load(map[string]string{"tradeStatus": "B001"}, "")
	if refundNotify.IsSuccess() {
		t.Fatal("refund notify should NOT succeed on tradeStatus != A001")
	}
}

func TestUnifiedOrderResponseTnFallsBackToUppercaseTN(t *testing.T) {
	var r JsapiUnifiedOrderResponse
	r.load(map[string]string{"TN": "legacy-tn"}, "")
	if r.Tn() != "legacy-tn" {
		t.Fatalf("Tn() = %s, want legacy-tn", r.Tn())
	}
	r.load(map[string]string{"tn": "lower-tn", "TN": "legacy-tn"}, "")
	if r.Tn() != "lower-tn" {
		t.Fatalf("Tn() = %s, want lower-tn (lowercase should take priority)", r.Tn())
	}
}

func TestParseResponseDecodesFormEncodedBody(t *testing.T) {
	got := parseResponse("a=1&b=hello%20world&empty=")
	if got["a"] != "1" || got["b"] != "hello world" || got["empty"] != "" {
		t.Fatalf("parseResponse() = %#v", got)
	}
	if len(parseResponse("")) != 0 {
		t.Fatal("parseResponse of empty string should return empty map")
	}
}
