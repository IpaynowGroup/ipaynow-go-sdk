package ipaynow

// ParsePaymentNotify parses and verifies an N001 payment callback body.
// N001 is a forward-transaction notification, so signType participates in
// the signature (excludeSignType=false), matching WP001/MQ002.
func ParsePaymentNotify(body string, config Config) (*PaymentNotifyResponse, error) {
	fields := parseResponse(body)
	if !Verify(fields, config, "signature", false) {
		return nil, &IpaynowError{Code: CodeSignError, Message: "现在支付回调通知验签失败"}
	}
	resp := &PaymentNotifyResponse{}
	resp.load(fields, body)
	return resp, nil
}

// ParseRefundNotify parses and verifies an RN001 refund callback body.
// RN001 belongs to the refund family, so signType is excluded from the
// signature (excludeSignType=true), matching R001/Q001/R002/Q002.
func ParseRefundNotify(body string, config Config) (*RefundNotifyResponse, error) {
	fields := parseResponse(body)
	if !Verify(fields, config, "signature", true) {
		return nil, &IpaynowError{Code: CodeSignError, Message: "现在支付回调通知验签失败"}
	}
	resp := &RefundNotifyResponse{}
	resp.load(fields, body)
	return resp, nil
}

// NotifyAckSuccess returns the synchronous ack body a merchant server must
// return to iPayNow after successfully handling a callback.
func NotifyAckSuccess() string { return "success=Y" }

// NotifyAckFail returns the synchronous ack body a merchant server must
// return to iPayNow after failing to handle a callback.
func NotifyAckFail() string { return "success=N" }
