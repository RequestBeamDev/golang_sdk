# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

用於 RequestBeam 文字、圖片、影片、Similarweb 和 Twitter API 的 Go 用戶端。支援 SSE 串流、有上限的工作輪詢和分頁，以及專案全部 12 種語言的 SDK 提示。

## 安裝與開始

需要 Go 1.22 或更新版本。在 Go 模組內執行安裝指令。沒有發佈標籤時，Go 使用虛擬版本；請保留 go.mod 中鎖定的版本。使用 ListModels(ctx) 查詢適合帳號與介面的模型；目錄中存在不代表供應商目前可用。

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("zh-TW"))
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

## 語言與回應

WithLocale 僅翻譯 SDK 自有提示，不翻譯 API 欄位、提示詞、模型 ID 或伺服器錯誤。支援下方全部語言，以及 pt、cn、zh、tw 等別名，並正規化底線與大小寫。未知語言退回英語。回應為 map[string]any，數字使用 json.Number 保留精度，ID 保留字串。預設位址為 https://api.requestbeam.com，不附加 /v1。自訂位址必須是受信任的 HTTPS 來源位址，SDK 禁止跟隨重新導向。

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## 串流與工作

StreamChatCompletions、StreamResponses、StreamMessages 和 StreamImage 將包含 Event、Data、ID 的 SSEEvent 傳給回呼。並非所有事件都包含文字。回呼傳回錯誤或取消 context 可停止讀取；[DONE] 由 SDK 內部處理。圖片直接傳回 JSON 或 SSE，不使用影片 Jobs。GenerateVideo 傳回 data.jobId。WaitForJob 在 COMPLETED 時結束；FAILED 和 CANCELLED 傳回 JobFailedError。逾時或取消只停止本機等待，不取消遠端工作或計費。

預設 60 秒逾時涵蓋整個回應，包括串流讀取。較長生成可設定 `requestbeam.WithTimeout(10*time.Minute)`，但呼叫端 context 或自訂 HTTP 用戶端的較早截止時間仍然有效。單一 SSE 事件上限為 64 MiB。`response.incomplete` 會結束串流讀取，但結果並未完整生成，請檢查狀態與 `incomplete_details`。

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb 與 Twitter

Similarweb 支援 27 個公開操作，只提供全球資料（country=999）；其他被接受的國家值也會傳回全球資料。需要日期時同時傳入 from 與 to，SDK 不修改日期或原始回應欄位。空結果不等於零流量。TwitterUserByUsername、TwitterUser 和 TwitterUserTweets 傳回原始回應。TwitterUserTweetPages 按正數頁數上限跟隨 meta.next_token，拒絕重複游標。每一頁都可能單獨收費。

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## 錯誤與驗證

APIError 保留 HTTP 狀態、伺服器錯誤碼、請求 ID 與原始回應。SDKError 表示本機錯誤，PollingTimeoutError 與 JobFailedError 區分等待逾時和工作失敗。SDK 本身不實作自動重試邏輯，包括 GET。網路失敗不代表伺服器沒有執行或計費。不要把金鑰放進 Git 或日誌；下載結果檔案時使用不帶認證的獨立用戶端。測試使用本機模擬服務；執行範例會傳送真實請求，可能產生費用。

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/tw)
