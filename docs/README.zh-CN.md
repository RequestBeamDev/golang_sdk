# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

用于 RequestBeam 文本、图片、视频、Similarweb 和 Twitter API 的 Go 客户端。支持 SSE 流式响应、有上限的任务轮询和分页，以及项目全部 12 种语言的 SDK 提示。

## 安装与开始

要求 Go 1.22 或更新版本。在 Go 模块内执行安装命令。没有发布标签时，Go 使用伪版本；请保留 go.mod 中锁定的版本。使用 ListModels(ctx) 查询适合账号和接口的模型；目录中存在不代表供应商当前可用。

```bash
go get github.com/RequestBeamDev/golang_sdk@latest
export REQUESTBEAM_API_KEY="YOUR_API_KEY"
export REQUESTBEAM_TEXT_MODEL="YOUR_TEXT_MODEL_ID"
```

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "os"

    requestbeam "github.com/RequestBeamDev/golang_sdk"
)

func main() {
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("zh-CN"))
    if err != nil {
        log.Fatal(err)
    }
    result, err := client.ChatCompletions(context.Background(), map[string]any{
        "model": os.Getenv("REQUESTBEAM_TEXT_MODEL"),
        "messages": []map[string]any{{"role": "user", "content": "Hello"}},
    })
    if err != nil {
        log.Fatal(err)
    }
    if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
        log.Fatal(err)
    }
}
```

## 语言与响应

WithLocale 仅翻译 SDK 自有提示，不翻译 API 字段、提示词、模型 ID 或服务端错误。支持下方全部语言，以及 pt、cn、zh、tw 等别名，并规范化下划线与大小写。未知语言回退为英语。响应为 map[string]any，数字使用 json.Number 保留精度，ID 保留字符串。默认地址为 https://api.requestbeam.com，不附加 /v1。自定义地址必须是受信任的 HTTPS 源地址，SDK 禁止跟随重定向。

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## 流式响应与任务

StreamChatCompletions、StreamResponses、StreamMessages 和 StreamImage 将包含 Event、Data、ID 的 SSEEvent 传给回调。并非所有事件都包含文本。回调返回错误或取消 context 可停止读取；[DONE] 由 SDK 内部处理。图片直接返回 JSON 或 SSE，不使用视频 Jobs。GenerateVideo 返回 data.jobId。WaitForJob 在 COMPLETED 时结束；FAILED 和 CANCELLED 返回 JobFailedError。超时或取消只停止本地等待，不取消远端任务或计费。

默认 60 秒超时覆盖整个响应，包括流式读取。较长生成可配置 `requestbeam.WithTimeout(10*time.Minute)`，但调用方 context 或自定义 HTTP 客户端的更早截止时间仍然有效。单个 SSE 事件上限为 64 MiB。`response.incomplete` 会结束流式读取，但结果并未完整生成，请检查状态与 `incomplete_details`。

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb 与 Twitter

Similarweb 支持 27 个公开操作，只提供全球数据（country=999）；其他被接受的国家值也会返回全球数据。需要日期时同时传入 from 与 to，SDK 不修改日期或原始响应字段。空结果不等于零流量。TwitterUserByUsername、TwitterUser 和 TwitterUserTweets 返回原始响应。TwitterUserTweetPages 按正数页数上限跟随 meta.next_token，拒绝重复游标。每一页都可能单独收费。

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## 错误与验证

APIError 保留 HTTP 状态、服务端错误码、请求 ID 和原始响应。SDKError 表示本地错误，PollingTimeoutError 与 JobFailedError 区分等待超时和任务失败。SDK 自身不实现自动重试逻辑，包括 GET。网络失败不代表服务端没有执行或计费。不要把密钥放进 Git 或日志；下载结果文件时使用不带认证的独立客户端。测试使用本地模拟服务；运行示例会发送真实请求，可能产生费用。

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/cn)
