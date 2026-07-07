# iPayNow Go SDK

现在支付 Go SDK，封装统一下单、订单查询、退款、撤销、回调验签，以及前端表单提交字段生成。

## 环境要求

- Go 1.21+

## 安装

```bash
go get github.com/IpaynowGroup/ipaynow-go-sdk
```

在代码中引入：

```go
import ipaynow "github.com/IpaynowGroup/ipaynow-go-sdk"
```

## 快速开始

以下是一个完整的 H5 统一下单示例：

```go
package main

import (
	"fmt"
	"log"
	"time"

	ipaynow "github.com/IpaynowGroup/ipaynow-go-sdk"
)

func main() {
	config, err := ipaynow.NewConfig("appId", "appKey")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	client := ipaynow.NewClient(config)

	params := ipaynow.H5UnifiedOrderParams{
		MhtOrderNo:        "ORDER202607030001",
		MhtOrderName:      "测试订单",
		MhtOrderDetail:    "测试订单",
		MhtOrderAmt:       1,
		MhtOrderTimeOut:   120,
		MhtOrderStartTime: time.Now().Format("20060102150405"),
		NotifyUrl:         "https://example.com/notify",
		FrontNotifyUrl:    "https://example.com/front",
		PayChannelType:    "12",
		OutputType:        "1",
		ConsumerCreateIp:  "127.0.0.1",
	}

	resp, err := ipaynow.Execute(client, ipaynow.NewH5UnifiedOrderRequest(params))
	if err != nil {
		log.Fatalf("execute error: %v", err)
	}
	if resp.IsSuccess() {
		fmt.Println("tn =", resp.Tn())
	} else {
		fmt.Println("failed:", resp.Code(), resp.Message())
	}
}
```

`Config` 的默认值：`BaseURL` 为 `https://pay.ipaynow.cn`，`ConnectTimeout` 为 5 秒，`ReadTimeout` 为 15 秒，均可通过 `Option`（`WithBaseURL`/`WithConnectTimeout`/`WithReadTimeout`）覆盖。

## 前端 Form 表单提交

需要让浏览器直接向现在支付网关提交表单时，使用 `BuildFormFields` 生成完整请求字段。入参与 `Execute` 一致，SDK 不发起 HTTP 请求：

```go
fields := ipaynow.BuildFormFields(client, ipaynow.NewH5UnifiedOrderRequest(params))
```

返回字段包含业务参数、公共参数和签名字段，例如：

- `funcode`
- `version`
- `appId`
- `mhtCharset`
- `mhtSignType` 或 `signType`
- `mhtSignature`

## 支持接口

| funcode | 说明 | 构造函数 |
| --- | --- | --- |
| `WP001` | JSAPI 统一下单 | `NewJsapiUnifiedOrderRequest` |
| `WP001` | 聚合码统一下单 | `NewAggregateCodeUnifiedOrderRequest` |
| `WP001` | 小程序统一下单 | `NewMiniProgramUnifiedOrderRequest` |
| `WP001` | H5 统一下单 | `NewH5UnifiedOrderRequest` |
| `WP001` | APP 统一下单 | `NewAppUnifiedOrderRequest` |
| `MQ002` | 订单查询 | `NewOrderQueryRequest` |
| `N001` | 支付结果回调 | `ParsePaymentNotify` |
| `R001` | 退款 | `NewRefundRequest` |
| `Q001` | 退款查询 | `NewRefundQueryRequest` |
| `RN001` | 退款结果回调 | `ParseRefundNotify` |
| `R002` | 撤销 | `NewReverseRequest` |
| `Q002` | 撤销查询 | `NewReverseQueryRequest` |

每个下单/查询/退款/撤销请求都提供了默认版本号的构造函数，以及可显式指定 `version` 的 `...WithVersion` 变体（`OrderQueryParams`/`ReverseQueryParams` 例外，通过 `Params.Version` 字段覆盖默认版本）。

## 响应判断

- 正向同步接口（统一下单、订单查询）默认使用 `Response.IsSuccess()`，规则为 `responseCode == "A001"`。
- 退款、撤销、退款查询、撤销查询覆盖了 `IsSuccess()`，规则为 `responseCode == "R000"`：

```go
func (r *RefundResponse) IsSuccess() bool { return r.Code() == "R000" }
func (r *RefundQueryResponse) IsSuccess() bool { return r.Code() == "R000" }
func (r *ReverseResponse) IsSuccess() bool { return r.Code() == "R000" }
func (r *ReverseQueryResponse) IsSuccess() bool { return r.Code() == "R000" }
```

- 支付回调 `PaymentNotifyResponse.IsSuccess()` 用 `transStatus == "A001"` 判断。
- 退款回调 `RefundNotifyResponse.IsSuccess()` 用 `tradeStatus == "A001"` 判断。

也可以直接读取原始字段：

```go
code := resp.Code()       // responseCode
message := resp.Message() // responseMsg
rawBody := resp.RawBody()
fields := resp.Fields()
value := resp.Get("nowPayOrderNo")
```

## 扩展字段

每个 `Params` 结构体都带有 `ExtraParams map[string]string` 字段。`ToMap()` 会先填充已知字段，再合并 `ExtraParams`（`mergeExtra` 只合并非空值，且会覆盖同名的已知字段），业务字段不做强校验，避免网关协议扩展时 SDK 阻塞接入：

```go
params := ipaynow.H5UnifiedOrderParams{
	MhtOrderNo:  "ORDER202607030001",
	MhtOrderAmt: 1,
	ExtraParams: map[string]string{
		"mhtGoodsTag": "goods-tag",
	},
}
```

注意：`Client.signedFields` 会在业务字段之上追加公共签名字段（`funcode`/`version`/`appId`/`mhtCharset`/签名类型字段/签名字段），并最终以这些公共字段为准，业务方不应通过 `ExtraParams` 覆盖它们：

- `funcode`
- `version`
- `appId`
- `mhtCharset`
- `mhtSignType`
- `signType`
- `mhtSignature`
- `signature`

## 回调验签

支付结果回调（N001）：

```go
notify, err := ipaynow.ParsePaymentNotify(body, appKey)
if err != nil {
	// 验签失败会返回 *IpaynowError，Code 为 ipaynow.CodeSignError
	http.Error(w, ipaynow.NotifyAckFail(), http.StatusOK)
	return
}
if notify.IsSuccess() {
	w.Write([]byte(ipaynow.NotifyAckSuccess()))
} else {
	w.Write([]byte(ipaynow.NotifyAckFail()))
}
```

退款结果回调（RN001）：

```go
notify, err := ipaynow.ParseRefundNotify(body, appKey)
```

`NotifyAckSuccess()`/`NotifyAckFail()` 分别返回 `"success=Y"`/`"success=N"`，是商户服务器同步响应回调所必须返回的内容。

## 错误处理

SDK 所有失败路径都返回同一个错误类型 `*IpaynowError`，调用方只需类型断言这一种类型并根据 `Code` 分支处理：

```go
resp, err := ipaynow.Execute(client, req)
if err != nil {
	var ipErr *ipaynow.IpaynowError
	if errors.As(err, &ipErr) {
		switch ipErr.Code {
		case ipaynow.CodeConnectTimeout:
			// 建连超时，可重试
		case ipaynow.CodeSocketTimeout:
			// 读取超时
		case ipaynow.CodeSignError:
			// 响应验签失败，绝不能当成功处理
		case ipaynow.CodeHTTPException:
			// 网络异常或非 2xx 响应
		case ipaynow.CodeSystemError:
			// SDK 内部错误
		}
	}
	return
}
```

错误码表：

| Code | 说明 |
| --- | --- |
| `E0001` | `CodeSystemError`，SDK 内部错误 |
| `E0002` | `CodeHTTPException`，连接失败或收到非 2xx 响应 |
| `E0003` | `CodeConnectTimeout`，建立 TCP 连接超时 |
| `E0004` | `CodeSocketTimeout`，等待/读取响应超时 |
| `E0012` | `CodeSignError`，响应或回调签名缺失/校验失败 |

`IpaynowError` 还实现了 `Unwrap() error`，可以用 `errors.Is`/`errors.As` 追溯底层原因（`Cause` 字段）。

## 日志

`Client` 默认使用 `slog.Default()`，可以通过 `WithLogger` 注入自定义的 `*slog.Logger`：

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
client := ipaynow.NewClient(config, ipaynow.WithLogger(logger))
```

请求/响应报文只在 `DEBUG` 级别输出，且 `mhtSignature`/`signature` 字段会被脱敏为 `***`；HTTP 层的连接/读取失败与超时在 `ERROR` 级别记录。`appKey` 不会出现在任何日志中。
