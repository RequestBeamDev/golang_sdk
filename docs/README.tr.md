# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

RequestBeam metin, görsel, video, Similarweb ve Twitter API'leri için Go istemcisi. SSE akışı, sınırlı iş sorgulama, sayfalama ve projenin 12 dilinde SDK mesajları sunar.

## Kurulum ve başlangıç

Go 1.22 veya üstü gerekir. Kurulum komutunu bir Go modülünde çalıştırın. Sürüm etiketi yoksa Go sözde sürüm kullanır; go.mod içindeki sürümü sabitleyin. Hesabınıza uygun modeli ListModels(ctx) ile seçin. Katalog kaydı sağlayıcı kullanılabilirliğini garanti etmez.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("tr"))
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

## Dil ve yanıtlar

WithLocale yalnızca SDK mesajlarını çevirir. API alanları, istemler, model kimlikleri ve sunucu mesajları değişmez. Aşağıdaki tüm diller desteklenir; pt, cn, zh, tw takma adları, alt çizgiler ve harf büyüklüğü normalleştirilir. Bilinmeyen dil İngilizceye döner. Yanıtlar map[string]any, sayılar hassasiyeti koruyan json.Number, kimlikler metindir. Varsayılan adres /v1 eklenmeden https://api.requestbeam.com şeklindedir. Özel adres güvenilir HTTPS kökeni olmalıdır; yönlendirmeler engellenir.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Akış ve işler

StreamChatCompletions, StreamResponses, StreamMessages ve StreamImage, Event, Data ve ID alanları bulunan SSEEvent değerlerini geri çağrıya iletir. Her olay metin içermez. Geri çağrı hatası veya context iptali okumayı durdurur; [DONE] içeride işlenir. Görseller doğrudan JSON veya SSE döndürür, video Jobs kullanmaz. GenerateVideo data.jobId döndürür. WaitForJob COMPLETED durumunda biter; FAILED ve CANCELLED, JobFailedError üretir. Süre aşımı veya context iptali yalnızca yerel beklemeyi durdurur, uzak işi veya ücretlendirmeyi iptal etmez.

Varsayılan 60 saniyelik süre, akış dahil tüm yanıtı kapsar. Uzun üretimler için `requestbeam.WithTimeout(10*time.Minute)` kullanın; daha erken context veya özel HTTP istemcisi süresi yine geçerlidir. Tek SSE olayı 64 MiB ile sınırlıdır. `response.incomplete` akışı bitirir ancak sonuç eksiktir; durumu ve `incomplete_details` alanını inceleyin.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb ve Twitter

Similarweb 27 işlem ve yalnızca dünya geneli veri sunar: country=999. Kabul edilen diğer ülke değerleri de dünya geneli veri verir. from ve to birlikte gönderilir; SDK tarihleri veya yerel yanıt alanlarını değiştirmez. Boş sonuç sıfır trafik demek değildir. TwitterUserByUsername, TwitterUser ve TwitterUserTweets yerel yanıtları döndürür. TwitterUserTweetPages pozitif sayfa sınırıyla meta.next_token izler ve tekrarlanan imleci reddeder. Her sayfa ücretli olabilir.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Hatalar ve doğrulama

APIError HTTP durumunu, hizmet kodunu, istek kimliğini ve özgün yanıtı korur. SDKError yerel hataları, PollingTimeoutError ve JobFailedError bekleme ve iş hatalarını belirtir. SDK'nin kendisi GET dahil hiçbir istek için otomatik yeniden deneme uygulamaz. Ağ hatası işlemin gerçekleşmediğini veya ücret alınmadığını kanıtlamaz. Anahtarları Git ve günlüklerden uzak tutun. Sonuç dosyalarını kimlik doğrulaması olmayan ayrı istemciyle indirin. Testler yerel taklitler kullanır; örnekler gerçek ve ücretli olabilecek istekler gönderir.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/tr)
