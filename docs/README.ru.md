# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

Go-клиент для API текста, изображений, видео, Similarweb и Twitter RequestBeam. Поддерживает SSE, ограниченный опрос заданий и пагинацию, а также сообщения SDK на всех 12 языках проекта.

## Установка и запуск

Требуется Go 1.22 или новее. Выполните команду установки внутри Go-модуля. Без тега выпуска Go использует псевдоверсию; сохраните версию в go.mod. Выберите подходящую модель через ListModels(ctx). Наличие в каталоге не гарантирует доступность провайдера.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("ru"))
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

## Язык и ответы

WithLocale переводит только сообщения SDK. Поля API, промпты, ID моделей и сообщения сервера не меняются. Поддерживаются все языки ниже; псевдонимы pt, cn, zh, tw, подчеркивания и регистр нормализуются. Неизвестный язык заменяется английским. Ответы имеют тип map[string]any, числа сохраняют точность через json.Number, ID остаются строками. Адрес по умолчанию — https://api.requestbeam.com без /v1. Собственный адрес должен быть доверенным HTTPS-источником; перенаправления заблокированы.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Потоки и задания

StreamChatCompletions, StreamResponses, StreamMessages и StreamImage передают в callback SSEEvent с Event, Data и ID. Не каждое событие содержит текст. Ошибка callback или отмена context останавливает чтение; [DONE] обрабатывается внутри SDK. Изображения возвращают JSON или SSE напрямую, без видео Jobs. GenerateVideo возвращает data.jobId. WaitForJob завершается при COMPLETED; FAILED и CANCELLED дают JobFailedError. Тайм-аут или отмена прекращают только локальное ожидание, но не удаленное задание и его оплату.

Стандартный тайм-аут 60 секунд охватывает весь ответ, включая поток. Для долгой генерации задайте `requestbeam.WithTimeout(10*time.Minute)`; более ранний срок context или HTTP-клиента сохраняет силу. Одно событие SSE ограничено 64 MiB. `response.incomplete` завершает поток, но результат неполный: проверьте статус и `incomplete_details`.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb и Twitter

Similarweb поддерживает 27 операций и только мировые данные: country=999. Другие допустимые страны также дают мировые данные. Передавайте from и to вместе; SDK не меняет даты и исходные поля. Пустой результат не доказывает нулевой трафик. TwitterUserByUsername, TwitterUser и TwitterUserTweets возвращают исходные ответы. TwitterUserTweetPages следует meta.next_token с положительным лимитом страниц и отклоняет повторные курсоры. Каждая страница может оплачиваться отдельно.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Ошибки и проверка

APIError сохраняет HTTP-статус, код сервиса, ID запроса и исходный ответ. SDKError обозначает локальные ошибки; PollingTimeoutError и JobFailedError различают истечение ожидания и сбой задания. Сам SDK не реализует автоматические повторы, даже для GET. Сетевая ошибка не исключает обработку и списание средств. Не помещайте ключи в Git или журналы. Загружайте файлы отдельным клиентом без авторизации. Тесты используют локальные имитации; примеры отправляют реальные запросы, которые могут быть платными.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/ru)
