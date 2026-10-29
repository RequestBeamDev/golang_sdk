# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

Go-Client für die Text-, Bild-, Video-, Similarweb- und Twitter-APIs von RequestBeam. Enthält SSE-Streaming, begrenztes Polling, Seitennavigation und SDK-Meldungen in allen 12 Projektsprachen.

## Installation und Start

Go 1.22 oder neuer ist erforderlich. Führen Sie den Installationsbefehl in einem Go-Modul aus. Ohne Release-Tag verwendet Go eine Pseudoversion; behalten Sie die Version in go.mod bei. Wählen Sie mit ListModels(ctx) ein passendes Modell für Ihr Konto. Ein Katalogeintrag garantiert keine Anbieter-Verfügbarkeit.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("de"))
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

## Sprache und Antworten

WithLocale übersetzt nur SDK-eigene Meldungen. API-Felder, Prompts, Modell-IDs und Servermeldungen bleiben unverändert. Unterstützt werden alle unten aufgeführten Sprachen; pt, cn, zh, tw sowie Unterstriche und Groß-/Kleinschreibung werden normalisiert. Unbekannte Sprachen fallen auf Englisch zurück. Antworten sind map[string]any; Zahlen bleiben als json.Number präzise, IDs bleiben Zeichenfolgen. Die Standardadresse ist https://api.requestbeam.com ohne /v1. Eigene Basisadressen müssen vertrauenswürdige HTTPS-Ursprünge sein; Weiterleitungen sind gesperrt.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Streaming und Aufträge

StreamChatCompletions, StreamResponses, StreamMessages und StreamImage liefern SSEEvent mit Event, Data und ID an einen Callback. Nicht jedes Ereignis enthält Text. Ein Callback-Fehler oder context-Abbruch beendet das Lesen; [DONE] wird intern verarbeitet. Bilder liefern direkt JSON oder SSE und verwenden keine Video-Jobs. GenerateVideo liefert data.jobId. WaitForJob endet bei COMPLETED; FAILED und CANCELLED liefern JobFailedError. Zeitüberschreitung und context-Abbruch stoppen nur das lokale Warten, nicht den Auftrag oder seine Abrechnung.

Das Standardzeitlimit von 60 Sekunden gilt für die gesamte Antwort einschließlich Streaming. Für längere Generierungen können Sie `requestbeam.WithTimeout(10*time.Minute)` setzen; eine frühere context-Frist oder das Zeitlimit des eigenen HTTP-Clients gilt weiterhin. Ein SSE-Ereignis ist auf 64 MiB begrenzt. `response.incomplete` beendet den Stream, bedeutet aber ein unvollständiges Ergebnis; prüfen Sie Status und `incomplete_details`.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb und Twitter

Similarweb unterstützt 27 Operationen und nur weltweite Daten: country=999. Andere akzeptierte Länderwerte liefern ebenfalls weltweite Daten. Übergeben Sie from und to gemeinsam; das SDK verändert keine Datumswerte oder nativen Antwortfelder. Leere Ergebnisse bedeuten nicht null Traffic. TwitterUserByUsername, TwitterUser und TwitterUserTweets liefern native Antworten. TwitterUserTweetPages folgt meta.next_token mit einer positiven Seitengrenze und weist wiederholte Cursor zurück. Jede Seite kann kostenpflichtig sein.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Fehler und Prüfung

APIError enthält HTTP-Status, Dienstcode, Request-ID und Originalantwort. SDKError beschreibt lokale Fehler; PollingTimeoutError und JobFailedError unterscheiden Wartezeit und Auftragsfehler. Das SDK selbst implementiert keine automatischen Wiederholungen, auch nicht für GETs. Ein Netzwerkfehler schließt Verarbeitung oder Abrechnung nicht aus. Halten Sie Schlüssel aus Git und Logs fern. Ergebnisdateien mit einem separaten Client ohne Authentifizierung laden. Tests nutzen lokale Mocks; Beispiele senden echte, möglicherweise kostenpflichtige Anfragen.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/de)
