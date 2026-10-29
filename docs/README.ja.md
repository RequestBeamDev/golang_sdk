# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

RequestBeam のテキスト、画像、動画、Similarweb、Twitter API 用 Go クライアントです。SSE、上限付きポーリングとページ分割、プロジェクトの全12言語の SDK メッセージに対応します。

## インストールと開始

Go 1.22 以降が必要です。Go モジュール内でインストールコマンドを実行してください。リリースタグがなければ疑似バージョンを使用するため、go.mod のバージョンを固定してください。ListModels(ctx) でアカウントに適したモデルを確認できますが、掲載はプロバイダーの稼働を保証しません。

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("ja"))
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

## 言語と応答

WithLocale は SDK 固有のメッセージのみを翻訳します。API フィールド、プロンプト、モデル ID、サーバーメッセージは変えません。以下の全言語に対応し、pt、cn、zh、tw、アンダースコア、大文字小文字を正規化します。未対応の言語は英語に戻ります。応答は map[string]any、数値は精度を保持する json.Number、ID は文字列です。既定の接続先は /v1 を付けない https://api.requestbeam.com です。独自 URL は信頼する HTTPS オリジンを指定し、リダイレクトは禁止されます。

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## ストリーミングとジョブ

StreamChatCompletions、StreamResponses、StreamMessages、StreamImage は Event、Data、ID を持つ SSEEvent をコールバックに渡します。すべてのイベントにテキストがあるわけではありません。コールバックのエラーや context のキャンセルで読み取りを終了し、[DONE] は内部で処理します。画像は JSON または SSE を直接返し、動画の Jobs は使いません。GenerateVideo は data.jobId を返します。WaitForJob は COMPLETED で完了し、FAILED と CANCELLED では JobFailedError を返します。タイムアウトやキャンセルはローカルの待機のみを止め、リモート処理や課金は取り消しません。

既定の60秒の制限はストリーミングを含む応答全体に適用されます。長い生成には `requestbeam.WithTimeout(10*time.Minute)` を指定できますが、context や独自 HTTP クライアントの短い期限が優先されます。SSE イベントは1件64 MiBまでです。`response.incomplete` はストリームの終了を示しますが、結果は未完了のため、ステータスと `incomplete_details` を確認してください。

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb と Twitter

Similarweb は27操作に対応し、データは全世界のみです（country=999）。他の受け付けられる国の値も全世界データになります。from と to は一緒に指定し、SDK は日付やネイティブフィールドを変更しません。空の結果はトラフィックゼロの証拠ではありません。TwitterUserByUsername、TwitterUser、TwitterUserTweets は元の応答を返します。TwitterUserTweetPages は正のページ上限で meta.next_token を追跡し、重複カーソルを拒否します。各ページに料金が発生する可能性があります。

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## エラーと検証

APIError は HTTP ステータス、サービスコード、リクエスト ID、元の応答を保持します。SDKError はローカルエラーを、PollingTimeoutError と JobFailedError は待機期限とジョブ失敗を区別します。SDK 自体は GET を含め自動再試行を実装しません。ネットワークエラーでも処理や課金が済んでいる可能性があります。キーを Git やログに残さず、結果ファイルは認証なしの別クライアントで取得してください。テストはローカルモックを使用しますが、サンプルは実際のリクエストを送り、課金される場合があります。

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/ja)
