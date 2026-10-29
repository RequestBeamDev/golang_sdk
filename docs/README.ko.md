# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

RequestBeam의 텍스트, 이미지, 동영상, Similarweb, Twitter API용 Go 클라이언트입니다. SSE, 제한된 작업 폴링과 페이지 처리, 프로젝트의 12개 언어로 된 SDK 메시지를 제공합니다.

## 설치와 시작

Go 1.22 이상이 필요합니다. Go 모듈 안에서 설치 명령을 실행하세요. 릴리스 태그가 없으면 Go는 의사 버전을 사용하므로 go.mod의 버전을 고정하세요. ListModels(ctx)로 계정에 적합한 모델을 확인하세요. 목록에 있다는 사실은 공급자 가용성을 보장하지 않습니다.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("ko"))
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

## 언어와 응답

WithLocale은 SDK 자체 메시지만 번역합니다. API 필드, 프롬프트, 모델 ID, 서버 메시지는 유지됩니다. 아래의 모든 언어를 지원하며 pt, cn, zh, tw 별칭, 밑줄, 대소문자를 정규화합니다. 알 수 없는 언어는 영어로 돌아갑니다. 응답은 map[string]any, 숫자는 정밀도를 유지하는 json.Number, ID는 문자열입니다. 기본 주소는 /v1 없는 https://api.requestbeam.com입니다. 사용자 지정 주소는 신뢰하는 HTTPS 오리진이어야 하며 리디렉션은 차단됩니다.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## 스트리밍과 작업

StreamChatCompletions, StreamResponses, StreamMessages, StreamImage는 Event, Data, ID를 가진 SSEEvent를 콜백에 전달합니다. 모든 이벤트에 텍스트가 있는 것은 아닙니다. 콜백 오류나 context 취소는 읽기를 중지하며 [DONE]은 내부 처리됩니다. 이미지는 JSON 또는 SSE를 직접 반환하고 동영상 Jobs를 사용하지 않습니다. GenerateVideo는 data.jobId를 반환합니다. WaitForJob는 COMPLETED에서 완료하며 FAILED와 CANCELLED는 JobFailedError를 반환합니다. 시간 초과와 취소는 로컬 대기만 중지하고 원격 작업이나 과금을 취소하지 않습니다.

기본 60초 제한은 스트리밍을 포함한 전체 응답에 적용됩니다. 긴 생성에는 `requestbeam.WithTimeout(10*time.Minute)`를 설정하세요. 더 짧은 context 또는 사용자 HTTP 클라이언트 제한은 계속 적용됩니다. SSE 이벤트는 하나당 64 MiB로 제한됩니다. `response.incomplete`는 스트림 종료지만 불완전한 결과이므로 상태와 `incomplete_details`를 확인하세요.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb과 Twitter

Similarweb은 27개 작업과 전 세계 데이터만 지원합니다(country=999). 허용되는 다른 국가 값도 전 세계 데이터를 반환합니다. from과 to를 함께 지정하며 SDK는 날짜나 원본 필드를 바꾸지 않습니다. 빈 결과는 트래픽이 0이라는 증거가 아닙니다. TwitterUserByUsername, TwitterUser, TwitterUserTweets는 원본 응답을 반환합니다. TwitterUserTweetPages는 양의 페이지 제한으로 meta.next_token을 따르고 반복 커서를 거부합니다. 각 페이지에 비용이 발생할 수 있습니다.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## 오류와 검증

APIError는 HTTP 상태, 서비스 코드, 요청 ID, 원본 응답을 보존합니다. SDKError는 로컬 오류를, PollingTimeoutError와 JobFailedError는 대기 만료와 작업 실패를 구분합니다. SDK 자체는 GET을 포함하여 자동 재시도를 구현하지 않습니다. 네트워크 오류가 처리나 과금이 없었음을 보장하지 않습니다. 키를 Git이나 로그에 남기지 말고 결과 파일은 인증 없는 별도 클라이언트로 받으세요. 테스트는 로컬 모의를 사용하지만 예제는 실제 요청을 보내므로 과금될 수 있습니다.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/ko)
