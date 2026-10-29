# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

Cliente Go para las API de texto, imágenes, vídeo, Similarweb y Twitter de RequestBeam. Incluye SSE, consulta de trabajos y paginación limitadas, y mensajes SDK en los 12 idiomas del proyecto.

## Instalación e inicio

Requiere Go 1.22 o posterior. Ejecuta el comando dentro de un módulo Go. Sin una etiqueta de versión, Go usa una pseudoversión; conserva la versión en go.mod. Selecciona con ListModels(ctx) un modelo adecuado para tu cuenta. El catálogo no garantiza la disponibilidad del proveedor.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("es"))
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

## Idioma y respuestas

WithLocale traduce solo los mensajes propios del SDK. Los campos API, prompts, ID de modelos y mensajes del servidor no cambian. Se admiten todos los idiomas de abajo; los alias pt, cn, zh, tw, los guiones bajos y las mayúsculas se normalizan. Un idioma desconocido usa inglés. Las respuestas son map[string]any; los números usan json.Number para conservar precisión y los ID son cadenas. La dirección predeterminada es https://api.requestbeam.com sin /v1. Las direcciones personalizadas deben ser orígenes HTTPS de confianza; no se siguen redirecciones.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Streaming y trabajos

StreamChatCompletions, StreamResponses, StreamMessages y StreamImage envían SSEEvent con Event, Data e ID al callback. No todos los eventos contienen texto. Un error del callback o cancelar el context detiene la lectura; [DONE] se procesa internamente. Las imágenes devuelven directamente JSON o SSE y no usan Jobs de vídeo. GenerateVideo devuelve data.jobId. WaitForJob termina con COMPLETED; FAILED y CANCELLED devuelven JobFailedError. El tiempo límite o la cancelación solo detienen la espera local, no el trabajo remoto ni la facturación.

El límite predeterminado de 60 segundos cubre toda la respuesta, incluido el streaming. Para generaciones largas usa `requestbeam.WithTimeout(10*time.Minute)`; un plazo anterior del context o del cliente HTTP sigue aplicándose. Cada evento SSE tiene un límite de 64 MiB. `response.incomplete` termina el flujo pero indica un resultado incompleto: revisa el estado e `incomplete_details`.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb y Twitter

Similarweb admite 27 operaciones y solo datos mundiales: country=999. Otros países aceptados también devuelven datos mundiales. Envía from y to juntos; el SDK no altera fechas ni campos nativos. Un resultado vacío no demuestra tráfico cero. TwitterUserByUsername, TwitterUser y TwitterUserTweets devuelven respuestas nativas. TwitterUserTweetPages sigue meta.next_token con un límite positivo de páginas y rechaza cursores repetidos. Cada página puede tener un coste.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Errores y validación

APIError conserva el estado HTTP, código del servicio, ID de solicitud y respuesta original. SDKError identifica errores locales; PollingTimeoutError y JobFailedError distinguen tiempo agotado y trabajo fallido. El propio SDK no implementa reintentos automáticos, tampoco para GET. Un error de red no descarta procesamiento o cobro. Mantén las claves fuera de Git y los registros. Descarga archivos resultantes con un cliente separado sin autenticación. Las pruebas usan simulaciones locales; los ejemplos envían solicitudes reales que pueden generar cargos.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/es)
