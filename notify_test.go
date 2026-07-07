package ipaynow

import "testing"

func TestParsePaymentNotifyExcludeSignTypeFalse(t *testing.T) {
	cfg := md5Config(t)
	fields := map[string]string{"transStatus": "A001", "signType": "MD5"}
	fields["signature"] = Sign(fields, cfg, false, "signature", "mhtSignature")
	body := "transStatus=A001&signType=MD5&signature=" + fields["signature"]

	resp, err := ParsePaymentNotify(body, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatal("expected payment notify to succeed")
	}
}

func TestParsePaymentNotifySignatureFailureReturnsSignError(t *testing.T) {
	body := "transStatus=A001&signature=deadbeef"
	_, err := ParsePaymentNotify(body, md5Config(t))
	assertSignError(t, err)
}

func TestParseRefundNotifyExcludeSignTypeTrue(t *testing.T) {
	cfg := md5Config(t)
	fields := map[string]string{"tradeStatus": "A001", "signType": "MD5"}
	fields["signature"] = Sign(fields, cfg, true, "signature", "mhtSignature")
	body := "tradeStatus=A001&signType=MD5&signature=" + fields["signature"]

	resp, err := ParseRefundNotify(body, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatal("expected refund notify to succeed")
	}

	// A signature computed with signType included must NOT verify, since
	// RN001 excludes signType.
	fieldsIncluded := map[string]string{"tradeStatus": "A001", "signType": "MD5"}
	fieldsIncluded["signature"] = Sign(fieldsIncluded, cfg, false, "signature", "mhtSignature")
	badBody := "tradeStatus=A001&signType=MD5&signature=" + fieldsIncluded["signature"]
	if _, err := ParseRefundNotify(badBody, cfg); err == nil {
		t.Fatal("expected sign error when signType wrongly participated in the signature")
	}
}

func TestParseRefundNotifyMissingSignatureFails(t *testing.T) {
	_, err := ParseRefundNotify("tradeStatus=A001", md5Config(t))
	assertSignError(t, err)
}

func TestNotifyAck(t *testing.T) {
	if NotifyAckSuccess() != "success=Y" {
		t.Fatalf("NotifyAckSuccess() = %s", NotifyAckSuccess())
	}
	if NotifyAckFail() != "success=N" {
		t.Fatalf("NotifyAckFail() = %s", NotifyAckFail())
	}
}

func assertSignError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	ipErr, ok := err.(*IpaynowError)
	if !ok {
		t.Fatalf("expected *IpaynowError, got %T", err)
	}
	if ipErr.Code != CodeSignError {
		t.Fatalf("Code = %s, want %s", ipErr.Code, CodeSignError)
	}
}
