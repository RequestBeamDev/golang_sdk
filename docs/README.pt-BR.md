# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

Cliente Go para as APIs de texto, imagem, vídeo, Similarweb e Twitter da RequestBeam. Inclui SSE, consulta de tarefas e paginação limitadas, com mensagens do SDK nos 12 idiomas do projeto.

## Instalação e início

Requer Go 1.22 ou superior. Execute o comando dentro de um módulo Go. Sem uma tag de versão, Go usa uma pseudoversão; mantenha a versão em go.mod. Use ListModels(ctx) para escolher um modelo adequado à sua conta. O catálogo não garante disponibilidade do provedor.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("pt-BR"))
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

## Idioma e respostas

WithLocale traduz apenas mensagens do SDK. Campos da API, prompts, IDs de modelos e mensagens do servidor permanecem intactos. Todos os idiomas abaixo são aceitos; aliases pt, cn, zh, tw, sublinhados e maiúsculas são normalizados. Um idioma desconhecido usa inglês. Respostas são map[string]any; números usam json.Number para preservar precisão, e IDs são strings. A raiz padrão é https://api.requestbeam.com sem /v1. Endereços personalizados devem ser origens HTTPS confiáveis; redirecionamentos são bloqueados.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Streaming e tarefas

StreamChatCompletions, StreamResponses, StreamMessages e StreamImage entregam SSEEvent com Event, Data e ID ao callback. Nem todo evento contém texto. Um erro do callback ou cancelamento do context interrompe a leitura; [DONE] é tratado internamente. Imagens retornam diretamente JSON ou SSE e não usam Jobs de vídeo. GenerateVideo retorna data.jobId. WaitForJob termina em COMPLETED; FAILED e CANCELLED retornam JobFailedError. Tempo limite ou cancelamento interrompe apenas a espera local, não a tarefa remota nem a cobrança.

O limite padrão de 60 segundos cobre toda a resposta, inclusive streaming. Para gerações longas use `requestbeam.WithTimeout(10*time.Minute)`; um prazo anterior do context ou do cliente HTTP continua valendo. Cada evento SSE tem limite de 64 MiB. `response.incomplete` encerra o fluxo, mas indica resultado incompleto: confira o status e `incomplete_details`.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb e Twitter

Similarweb aceita 27 operações e apenas dados mundiais: country=999. Outros países aceitos também retornam dados mundiais. Envie from e to juntos; o SDK não altera datas nem campos nativos. Resultados vazios não provam tráfego zero. TwitterUserByUsername, TwitterUser e TwitterUserTweets retornam respostas nativas. TwitterUserTweetPages segue meta.next_token com limite positivo de páginas e rejeita cursores repetidos. Cada página pode ter cobrança.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Erros e validação

APIError preserva status HTTP, código do serviço, ID da requisição e resposta original. SDKError identifica erros locais; PollingTimeoutError e JobFailedError distinguem prazo vencido e falha de tarefa. O próprio SDK não implementa novas tentativas automáticas, nem para GET. Um erro de rede não exclui processamento ou cobrança. Mantenha chaves fora do Git e dos logs. Baixe arquivos com outro cliente sem autenticação. Testes usam simulações locais; exemplos enviam requisições reais que podem gerar custos.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/pt)
