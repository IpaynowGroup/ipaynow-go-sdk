package ipaynow

import "strconv"

// --- field-mapping helpers shared by every Params.ToMap() -----------------
//
// "Zero value" for a string field means "not set" and is not sent (spec
// §5). Fields with a documented default are filled in here so callers only
// need to set a non-zero value to override the default explicitly.

func setStr(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	}
}

func setStrDefault(m map[string]string, key, value, def string) {
	if value != "" {
		m[key] = value
		return
	}
	if def != "" {
		m[key] = def
	}
}

func setInt64Required(m map[string]string, key string, value int64) {
	m[key] = strconv.FormatInt(value, 10)
}

func setInt64Default(m map[string]string, key string, value, def int64) {
	if value != 0 {
		m[key] = strconv.FormatInt(value, 10)
		return
	}
	if def != 0 {
		m[key] = strconv.FormatInt(def, 10)
	}
}

func setInt64Ptr(m map[string]string, key string, value *int64) {
	if value != nil {
		m[key] = strconv.FormatInt(*value, 10)
	}
}

// mergeExtra merges ExtraParams last so it can override a named field's
// value, matching the Java BeanMap behavior referenced in spec §5.
func mergeExtra(m map[string]string, extra map[string]string) {
	for k, v := range extra {
		if v != "" {
			m[k] = v
		}
	}
}

// JsapiUnifiedOrderParams is the WP001 JSAPI unified-order request body
// (spec §5.1).
type JsapiUnifiedOrderParams struct {
	ConsumerCreateIp  string
	ConsumerId        string
	DeviceType        string // default "0600"
	FrontNotifyUrl    string
	MhtCurrencyType   string // default "156"
	MhtGoodsTag       string
	MhtGoodsDetails   string
	MhtLimitPay       string
	MhtOrderAmt       int64 // required, amount in cents
	MhtOrderDetail    string
	MhtOrderName      string
	MhtOrderNo        string // required
	MhtOrderStartTime string
	MhtOrderTimeOut   int64  // default 120
	MhtOrderType      string // default "05"
	MhtReserved       string
	MhtSubAppId       string
	NotifyUrl         string
	OutputType        string
	PayChannelType    string
	ExtraParams       map[string]string
}

func (p JsapiUnifiedOrderParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "consumerCreateIp", p.ConsumerCreateIp)
	setStr(m, "consumerId", p.ConsumerId)
	setStrDefault(m, "deviceType", p.DeviceType, "0600")
	setStr(m, "frontNotifyUrl", p.FrontNotifyUrl)
	setStrDefault(m, "mhtCurrencyType", p.MhtCurrencyType, "156")
	setStr(m, "mhtGoodsTag", p.MhtGoodsTag)
	setStr(m, "mhtGoodsDetails", p.MhtGoodsDetails)
	setStr(m, "mhtLimitPay", p.MhtLimitPay)
	setInt64Required(m, "mhtOrderAmt", p.MhtOrderAmt)
	setStr(m, "mhtOrderDetail", p.MhtOrderDetail)
	setStr(m, "mhtOrderName", p.MhtOrderName)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "mhtOrderStartTime", p.MhtOrderStartTime)
	setInt64Default(m, "mhtOrderTimeOut", p.MhtOrderTimeOut, 120)
	setStrDefault(m, "mhtOrderType", p.MhtOrderType, "05")
	setStr(m, "mhtReserved", p.MhtReserved)
	setStr(m, "mhtSubAppId", p.MhtSubAppId)
	setStr(m, "notifyUrl", p.NotifyUrl)
	setStr(m, "outputType", p.OutputType)
	setStr(m, "payChannelType", p.PayChannelType)
	mergeExtra(m, p.ExtraParams)
	return m
}

// AggregateCodeUnifiedOrderParams is the WP001 aggregate-code unified-order
// request body (spec §5.2).
type AggregateCodeUnifiedOrderParams struct {
	DeviceType        string // default "20"
	MhtCurrencyType   string // default "156"
	MhtLimitPay       string
	MhtOrderAmt       int64 // required
	MhtOrderDetail    string
	MhtOrderName      string
	MhtOrderNo        string // required
	MhtOrderStartTime string
	MhtOrderTimeOut   int64  // default 120
	MhtOrderType      string // default "05"
	MhtReserved       string
	NotifyUrl         string
	OutputType        string
	ExtraParams       map[string]string
}

func (p AggregateCodeUnifiedOrderParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStrDefault(m, "deviceType", p.DeviceType, "20")
	setStrDefault(m, "mhtCurrencyType", p.MhtCurrencyType, "156")
	setStr(m, "mhtLimitPay", p.MhtLimitPay)
	setInt64Required(m, "mhtOrderAmt", p.MhtOrderAmt)
	setStr(m, "mhtOrderDetail", p.MhtOrderDetail)
	setStr(m, "mhtOrderName", p.MhtOrderName)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "mhtOrderStartTime", p.MhtOrderStartTime)
	setInt64Default(m, "mhtOrderTimeOut", p.MhtOrderTimeOut, 120)
	setStrDefault(m, "mhtOrderType", p.MhtOrderType, "05")
	setStr(m, "mhtReserved", p.MhtReserved)
	setStr(m, "notifyUrl", p.NotifyUrl)
	setStr(m, "outputType", p.OutputType)
	mergeExtra(m, p.ExtraParams)
	return m
}

// MiniProgramUnifiedOrderParams is the WP001 mini-program unified-order
// request body (spec §5.3).
type MiniProgramUnifiedOrderParams struct {
	ConsumerId        string
	DeviceType        string // default "14"
	DiscountAmt       *int64
	MhtCurrencyType   string // default "156"
	MhtGoodsTag       string
	MhtGoodsDetails   string
	MhtLimitPay       string
	MhtOrderAmt       int64 // required
	MhtOrderDetail    string
	MhtOrderName      string
	MhtOrderNo        string // required
	MhtOrderStartTime string
	MhtOrderTimeOut   int64  // default 120
	MhtOrderType      string // default "05"
	MhtReserved       string
	MhtSubAppId       string
	NotifyUrl         string
	OriMhtOrderAmt    *int64
	OutputType        string // default "1"
	PayChannelType    string
	MhtStoreId        string
	ExtraParams       map[string]string
}

func (p MiniProgramUnifiedOrderParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "consumerId", p.ConsumerId)
	setStrDefault(m, "deviceType", p.DeviceType, "14")
	setInt64Ptr(m, "discountAmt", p.DiscountAmt)
	setStrDefault(m, "mhtCurrencyType", p.MhtCurrencyType, "156")
	setStr(m, "mhtGoodsTag", p.MhtGoodsTag)
	setStr(m, "mhtGoodsDetails", p.MhtGoodsDetails)
	setStr(m, "mhtLimitPay", p.MhtLimitPay)
	setInt64Required(m, "mhtOrderAmt", p.MhtOrderAmt)
	setStr(m, "mhtOrderDetail", p.MhtOrderDetail)
	setStr(m, "mhtOrderName", p.MhtOrderName)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "mhtOrderStartTime", p.MhtOrderStartTime)
	setInt64Default(m, "mhtOrderTimeOut", p.MhtOrderTimeOut, 120)
	setStrDefault(m, "mhtOrderType", p.MhtOrderType, "05")
	setStr(m, "mhtReserved", p.MhtReserved)
	setStr(m, "mhtSubAppId", p.MhtSubAppId)
	setStr(m, "notifyUrl", p.NotifyUrl)
	setInt64Ptr(m, "oriMhtOrderAmt", p.OriMhtOrderAmt)
	setStrDefault(m, "outputType", p.OutputType, "1")
	setStr(m, "payChannelType", p.PayChannelType)
	setStr(m, "mhtStoreId", p.MhtStoreId)
	mergeExtra(m, p.ExtraParams)
	return m
}

// H5UnifiedOrderParams is the WP001 H5 unified-order request body
// (spec §5.4).
type H5UnifiedOrderParams struct {
	ConsumerCreateIp  string
	ConsumerId        string
	DeviceType        string // default "0601"
	FrontNotifyUrl    string
	MhtCurrencyType   string // default "156"
	MhtLimitPay       string
	MhtOrderAmt       int64 // required
	MhtOrderDetail    string
	MhtOrderName      string
	MhtOrderNo        string // required
	MhtOrderStartTime string
	MhtOrderTimeOut   int64  // default 120
	MhtOrderType      string // default "05"
	MhtReserved       string
	NotifyUrl         string
	OutputType        string
	PayChannelType    string
	SpecifiedChannel  string
	PayAccNo          string
	GuideId           string
	ExtraParams       map[string]string
}

func (p H5UnifiedOrderParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "consumerCreateIp", p.ConsumerCreateIp)
	setStr(m, "consumerId", p.ConsumerId)
	setStrDefault(m, "deviceType", p.DeviceType, "0601")
	setStr(m, "frontNotifyUrl", p.FrontNotifyUrl)
	setStrDefault(m, "mhtCurrencyType", p.MhtCurrencyType, "156")
	setStr(m, "mhtLimitPay", p.MhtLimitPay)
	setInt64Required(m, "mhtOrderAmt", p.MhtOrderAmt)
	setStr(m, "mhtOrderDetail", p.MhtOrderDetail)
	setStr(m, "mhtOrderName", p.MhtOrderName)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "mhtOrderStartTime", p.MhtOrderStartTime)
	setInt64Default(m, "mhtOrderTimeOut", p.MhtOrderTimeOut, 120)
	setStrDefault(m, "mhtOrderType", p.MhtOrderType, "05")
	setStr(m, "mhtReserved", p.MhtReserved)
	setStr(m, "notifyUrl", p.NotifyUrl)
	setStr(m, "outputType", p.OutputType)
	setStr(m, "payChannelType", p.PayChannelType)
	setStr(m, "specifiedChannel", p.SpecifiedChannel)
	setStr(m, "payAccNo", p.PayAccNo)
	setStr(m, "guideId", p.GuideId)
	mergeExtra(m, p.ExtraParams)
	return m
}

// AppUnifiedOrderParams is the WP001 APP unified-order request body
// (spec §5.5).
type AppUnifiedOrderParams struct {
	ConsumerCreateIp  string
	ConsumerId        string
	ConsumerName      string
	DeviceType        string // default "01"
	MhtCurrencyType   string // default "156"
	MhtLimitPay       string
	MhtOrderAmt       int64 // required
	MhtOrderDetail    string
	MhtOrderName      string
	MhtOrderNo        string // required
	MhtOrderStartTime string
	MhtOrderTimeOut   int64  // default 120
	MhtOrderType      string // default "05"
	MhtReserved       string
	MhtSubAppId       string
	NotifyUrl         string
	PayChannelType    string
	OutputType        string // default "0"
	FrontNotifyUrl    string
	MhtStoreId        string
	ExtraParams       map[string]string
}

func (p AppUnifiedOrderParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "consumerCreateIp", p.ConsumerCreateIp)
	setStr(m, "consumerId", p.ConsumerId)
	setStr(m, "consumerName", p.ConsumerName)
	setStrDefault(m, "deviceType", p.DeviceType, "01")
	setStrDefault(m, "mhtCurrencyType", p.MhtCurrencyType, "156")
	setStr(m, "mhtLimitPay", p.MhtLimitPay)
	setInt64Required(m, "mhtOrderAmt", p.MhtOrderAmt)
	setStr(m, "mhtOrderDetail", p.MhtOrderDetail)
	setStr(m, "mhtOrderName", p.MhtOrderName)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "mhtOrderStartTime", p.MhtOrderStartTime)
	setInt64Default(m, "mhtOrderTimeOut", p.MhtOrderTimeOut, 120)
	setStrDefault(m, "mhtOrderType", p.MhtOrderType, "05")
	setStr(m, "mhtReserved", p.MhtReserved)
	setStr(m, "mhtSubAppId", p.MhtSubAppId)
	setStr(m, "notifyUrl", p.NotifyUrl)
	setStr(m, "payChannelType", p.PayChannelType)
	setStrDefault(m, "outputType", p.OutputType, "0")
	setStr(m, "frontNotifyUrl", p.FrontNotifyUrl)
	setStr(m, "mhtStoreId", p.MhtStoreId)
	mergeExtra(m, p.ExtraParams)
	return m
}

// OrderQueryParams is the MQ002 order-query request body (spec §5.6).
// Version, when set, overrides the request's default version (1.0.0).
type OrderQueryParams struct {
	DeviceType  string
	MhtOrderNo  string // required
	Version     string
	ExtraParams map[string]string
}

func (p OrderQueryParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "deviceType", p.DeviceType)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	mergeExtra(m, p.ExtraParams)
	return m
}

// RefundParams is the R001 refund request body (spec §5.7).
type RefundParams struct {
	Amount      int64  // required, amount in cents, <= original order amount
	MhtOrderNo  string // required
	MhtRefundNo string // required
	NotifyUrl   string
	OriTransId  string
	Reason      string
	ExtraParams map[string]string
}

func (p RefundParams) ToMap() map[string]string {
	m := make(map[string]string)
	setInt64Required(m, "amount", p.Amount)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "mhtRefundNo", p.MhtRefundNo)
	setStr(m, "notifyUrl", p.NotifyUrl)
	setStr(m, "oriTransId", p.OriTransId)
	setStr(m, "reason", p.Reason)
	mergeExtra(m, p.ExtraParams)
	return m
}

// RefundQueryParams is the Q001 refund-query request body (spec §5.8).
type RefundQueryParams struct {
	MhtRefundNo string // required
	ExtraParams map[string]string
}

func (p RefundQueryParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "mhtRefundNo", p.MhtRefundNo)
	mergeExtra(m, p.ExtraParams)
	return m
}

// ReverseParams is the R002 reversal request body (spec §5.9).
type ReverseParams struct {
	MhtRefundNo string // required
	MhtOrderNo  string // required
	OriTransId  string
	Reason      string
	ExtraParams map[string]string
}

func (p ReverseParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "mhtRefundNo", p.MhtRefundNo)
	setStr(m, "mhtOrderNo", p.MhtOrderNo)
	setStr(m, "oriTransId", p.OriTransId)
	setStr(m, "reason", p.Reason)
	mergeExtra(m, p.ExtraParams)
	return m
}

// ReverseQueryParams is the Q002 reversal-query request body (spec §5.10).
type ReverseQueryParams struct {
	MhtRefundNo string // required
	ExtraParams map[string]string
}

func (p ReverseQueryParams) ToMap() map[string]string {
	m := make(map[string]string)
	setStr(m, "mhtRefundNo", p.MhtRefundNo)
	mergeExtra(m, p.ExtraParams)
	return m
}
