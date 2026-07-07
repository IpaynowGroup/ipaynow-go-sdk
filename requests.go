package ipaynow

// RequestMeta carries the wire-protocol metadata for one operation
// (spec §2): funcode/version/path plus the signature field names and
// whether signType participates in the signature for this operation.
type RequestMeta struct {
	Funcode                string
	Version                string
	MhtCharset             string
	Path                   string
	SignTypeField          string
	SignatureField         string
	ResponseSignatureField string
	ExcludeSignType        bool
	Params                 map[string]string
}

// Request pairs RequestMeta with a hydrate function that builds the
// operation's typed Response from parsed fields. Each operation gets its
// own NewXxxRequest constructor (a centralized factory was rejected in
// design review, see spec §1) rather than a generic map-in/map-out API.
type Request[T any] struct {
	RequestMeta
	newResponse func(fields map[string]string, rawBody string) T
}

// responseLoader is implemented by *Response and, through embedding, by
// every operation-specific response type in this package.
type responseLoader interface {
	load(fields map[string]string, rawBody string)
}

// newHydrate builds the hydrate closure stored on Request[T]. PT is
// constrained to *T so the compiler verifies T's pointer type actually
// implements responseLoader (true for every Response-embedding type here).
func newHydrate[T any, PT interface {
	*T
	responseLoader
}]() func(map[string]string, string) T {
	return func(fields map[string]string, rawBody string) T {
		var resp T
		PT(&resp).load(fields, rawBody)
		return resp
	}
}

// --- JSAPI unified order (WP001) -------------------------------------------

func NewJsapiUnifiedOrderRequest(params JsapiUnifiedOrderParams) *Request[JsapiUnifiedOrderResponse] {
	return NewJsapiUnifiedOrderRequestWithVersion("1.0.0", params)
}

func NewJsapiUnifiedOrderRequestWithVersion(version string, params JsapiUnifiedOrderParams) *Request[JsapiUnifiedOrderResponse] {
	return &Request[JsapiUnifiedOrderResponse]{
		RequestMeta: RequestMeta{
			Funcode: "WP001", Version: version, MhtCharset: "UTF-8", Path: "/",
			SignTypeField: "mhtSignType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: false, Params: params.ToMap(),
		},
		newResponse: newHydrate[JsapiUnifiedOrderResponse, *JsapiUnifiedOrderResponse](),
	}
}

// --- Aggregate code unified order (WP001) ----------------------------------

func NewAggregateCodeUnifiedOrderRequest(params AggregateCodeUnifiedOrderParams) *Request[AggregateCodeUnifiedOrderResponse] {
	return NewAggregateCodeUnifiedOrderRequestWithVersion("1.0.3", params)
}

func NewAggregateCodeUnifiedOrderRequestWithVersion(version string, params AggregateCodeUnifiedOrderParams) *Request[AggregateCodeUnifiedOrderResponse] {
	return &Request[AggregateCodeUnifiedOrderResponse]{
		RequestMeta: RequestMeta{
			Funcode: "WP001", Version: version, MhtCharset: "UTF-8", Path: "/",
			SignTypeField: "mhtSignType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: false, Params: params.ToMap(),
		},
		newResponse: newHydrate[AggregateCodeUnifiedOrderResponse, *AggregateCodeUnifiedOrderResponse](),
	}
}

// --- Mini program unified order (WP001) ------------------------------------

func NewMiniProgramUnifiedOrderRequest(params MiniProgramUnifiedOrderParams) *Request[MiniProgramUnifiedOrderResponse] {
	return NewMiniProgramUnifiedOrderRequestWithVersion("1.0.0", params)
}

func NewMiniProgramUnifiedOrderRequestWithVersion(version string, params MiniProgramUnifiedOrderParams) *Request[MiniProgramUnifiedOrderResponse] {
	return &Request[MiniProgramUnifiedOrderResponse]{
		RequestMeta: RequestMeta{
			Funcode: "WP001", Version: version, MhtCharset: "UTF-8", Path: "/",
			SignTypeField: "mhtSignType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: false, Params: params.ToMap(),
		},
		newResponse: newHydrate[MiniProgramUnifiedOrderResponse, *MiniProgramUnifiedOrderResponse](),
	}
}

// --- H5 unified order (WP001) ----------------------------------------------

func NewH5UnifiedOrderRequest(params H5UnifiedOrderParams) *Request[H5UnifiedOrderResponse] {
	return NewH5UnifiedOrderRequestWithVersion("1.0.0", params)
}

func NewH5UnifiedOrderRequestWithVersion(version string, params H5UnifiedOrderParams) *Request[H5UnifiedOrderResponse] {
	return &Request[H5UnifiedOrderResponse]{
		RequestMeta: RequestMeta{
			Funcode: "WP001", Version: version, MhtCharset: "UTF-8", Path: "/",
			SignTypeField: "mhtSignType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: false, Params: params.ToMap(),
		},
		newResponse: newHydrate[H5UnifiedOrderResponse, *H5UnifiedOrderResponse](),
	}
}

// --- APP unified order (WP001) ----------------------------------------------

func NewAppUnifiedOrderRequest(params AppUnifiedOrderParams) *Request[AppUnifiedOrderResponse] {
	return NewAppUnifiedOrderRequestWithVersion("1.0.3", params)
}

func NewAppUnifiedOrderRequestWithVersion(version string, params AppUnifiedOrderParams) *Request[AppUnifiedOrderResponse] {
	return &Request[AppUnifiedOrderResponse]{
		RequestMeta: RequestMeta{
			Funcode: "WP001", Version: version, MhtCharset: "UTF-8", Path: "/",
			SignTypeField: "mhtSignType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: false, Params: params.ToMap(),
		},
		newResponse: newHydrate[AppUnifiedOrderResponse, *AppUnifiedOrderResponse](),
	}
}

// --- Order query (MQ002) ----------------------------------------------------

// NewOrderQueryRequest builds the order-query request. If params.Version is
// set it overrides the default "1.0.0", matching the Java SDK's behavior of
// reading the version from the params rather than a constructor overload.
func NewOrderQueryRequest(params OrderQueryParams) *Request[OrderQueryResponse] {
	version := params.Version
	if version == "" {
		version = "1.0.0"
	}
	return &Request[OrderQueryResponse]{
		RequestMeta: RequestMeta{
			Funcode: "MQ002", Version: version, MhtCharset: "UTF-8", Path: "/",
			SignTypeField: "mhtSignType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: false, Params: params.ToMap(),
		},
		newResponse: newHydrate[OrderQueryResponse, *OrderQueryResponse](),
	}
}

// --- Refund (R001) -----------------------------------------------------------

func NewRefundRequest(params RefundParams) *Request[RefundResponse] {
	return NewRefundRequestWithVersion("1.0.0", params)
}

func NewRefundRequestWithVersion(version string, params RefundParams) *Request[RefundResponse] {
	return &Request[RefundResponse]{
		RequestMeta: RequestMeta{
			Funcode: "R001", Version: version, MhtCharset: "UTF-8", Path: "/refund/refundOrder",
			SignTypeField: "signType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: true, Params: params.ToMap(),
		},
		newResponse: newHydrate[RefundResponse, *RefundResponse](),
	}
}

// --- Refund query (Q001) ------------------------------------------------------

func NewRefundQueryRequest(params RefundQueryParams) *Request[RefundQueryResponse] {
	return NewRefundQueryRequestWithVersion("1.0.0", params)
}

func NewRefundQueryRequestWithVersion(version string, params RefundQueryParams) *Request[RefundQueryResponse] {
	return &Request[RefundQueryResponse]{
		RequestMeta: RequestMeta{
			Funcode: "Q001", Version: version, MhtCharset: "UTF-8", Path: "/refund/refundQuery",
			SignTypeField: "signType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: true, Params: params.ToMap(),
		},
		newResponse: newHydrate[RefundQueryResponse, *RefundQueryResponse](),
	}
}

// --- Reverse (R002) -----------------------------------------------------------

func NewReverseRequest(params ReverseParams) *Request[ReverseResponse] {
	return &Request[ReverseResponse]{
		RequestMeta: RequestMeta{
			Funcode: "R002", Version: "1.0.0", MhtCharset: "UTF-8", Path: "/refund/cancelOrder",
			SignTypeField: "signType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: true, Params: params.ToMap(),
		},
		newResponse: newHydrate[ReverseResponse, *ReverseResponse](),
	}
}

// --- Reverse query (Q002) ------------------------------------------------------

func NewReverseQueryRequest(params ReverseQueryParams) *Request[ReverseQueryResponse] {
	return NewReverseQueryRequestWithVersion("1.0.0", params)
}

func NewReverseQueryRequestWithVersion(version string, params ReverseQueryParams) *Request[ReverseQueryResponse] {
	return &Request[ReverseQueryResponse]{
		RequestMeta: RequestMeta{
			Funcode: "Q002", Version: version, MhtCharset: "UTF-8", Path: "/refund/refundQuery",
			SignTypeField: "signType", SignatureField: "mhtSignature", ResponseSignatureField: "signature",
			ExcludeSignType: true, Params: params.ToMap(),
		},
		newResponse: newHydrate[ReverseQueryResponse, *ReverseQueryResponse](),
	}
}
