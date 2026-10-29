# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

Client Go per le API di testo, immagini, video, Similarweb e Twitter di RequestBeam. Include SSE, polling e paginazione limitati e messaggi SDK nelle 12 lingue del progetto.

## Installazione e avvio

Richiede Go 1.22 o successivo. Esegui il comando in un modulo Go. Senza tag di rilascio Go usa una pseudo-versione; conserva la versione in go.mod. Scegli con ListModels(ctx) un modello adatto al tuo account. Il catalogo non garantisce la disponibilità del fornitore.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("it"))
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

## Lingua e risposte

WithLocale traduce solo i messaggi SDK. Campi API, prompt, ID dei modelli e messaggi del server restano invariati. Sono supportate tutte le lingue elencate sotto; alias pt, cn, zh, tw, trattini bassi e maiuscole vengono normalizzati. Una lingua sconosciuta usa l'inglese. Le risposte sono map[string]any, i numeri json.Number per mantenere la precisione e gli ID stringhe. L'indirizzo predefinito è https://api.requestbeam.com senza /v1. Gli indirizzi personalizzati devono essere origini HTTPS attendibili; i reindirizzamenti sono bloccati.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Streaming e processi

StreamChatCompletions, StreamResponses, StreamMessages e StreamImage passano al callback SSEEvent con Event, Data e ID. Non ogni evento contiene testo. Un errore del callback o l'annullamento del context interrompe la lettura; [DONE] viene gestito internamente. Le immagini restituiscono direttamente JSON o SSE, senza Jobs video. GenerateVideo restituisce data.jobId. WaitForJob termina con COMPLETED; FAILED e CANCELLED producono JobFailedError. Il timeout o l'annullamento interrompe solo l'attesa locale, non il processo remoto né la fatturazione.

Il timeout predefinito di 60 secondi copre l'intera risposta, streaming incluso. Per generazioni lunghe usa `requestbeam.WithTimeout(10*time.Minute)`; una scadenza precedente del context o del client HTTP resta valida. Ogni evento SSE è limitato a 64 MiB. `response.incomplete` termina il flusso ma indica un risultato incompleto: controlla stato e `incomplete_details`.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb e Twitter

Similarweb supporta 27 operazioni e solo dati mondiali: country=999. Anche altri valori di paese accettati restituiscono dati mondiali. Passa from e to insieme; l'SDK non modifica date o campi nativi. Risultati vuoti non dimostrano traffico nullo. TwitterUserByUsername, TwitterUser e TwitterUserTweets restituiscono risposte native. TwitterUserTweetPages segue meta.next_token con un limite positivo di pagine e rifiuta cursori ripetuti. Ogni pagina può essere a pagamento.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Errori e verifica

APIError conserva stato HTTP, codice del servizio, ID richiesta e risposta originale. SDKError identifica errori locali; PollingTimeoutError e JobFailedError distinguono scadenza e fallimento del processo. Il codice SDK non implementa tentativi automatici, neppure per GET. Un errore di rete non esclude elaborazione o addebito. Non inserire chiavi in Git o nei log. Scarica i file risultanti con un client separato senza autenticazione. I test usano simulazioni locali; gli esempi inviano richieste reali potenzialmente a pagamento.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/it)
