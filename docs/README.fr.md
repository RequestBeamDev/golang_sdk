# RequestBeam Go SDK

[English](../README.md) | [Deutsch](README.de.md) | [Türkçe](README.tr.md) | [Italiano](README.it.md) | [Español](README.es.md) | [Français](README.fr.md) | [Português](README.pt-BR.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-TW.md) | [Русский](README.ru.md)

Client Go pour les API de texte, images, vidéo, Similarweb et Twitter de RequestBeam. Inclut SSE, interrogation et pagination limitées, ainsi que des messages SDK dans les 12 langues du projet.

## Installation et démarrage

Go 1.22 ou ultérieur est requis. Exécutez la commande dans un module Go. Sans tag de version, Go utilise une pseudo-version ; conservez celle de go.mod. Choisissez avec ListModels(ctx) un modèle adapté à votre compte. Le catalogue ne garantit pas la disponibilité du fournisseur.

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
    client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("fr"))
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

## Langue et réponses

WithLocale traduit seulement les messages du SDK. Les champs API, prompts, ID de modèles et messages du serveur restent inchangés. Toutes les langues ci-dessous sont prises en charge ; les alias pt, cn, zh, tw, traits de soulignement et majuscules sont normalisés. Une langue inconnue utilise l'anglais. Les réponses sont des map[string]any ; json.Number préserve la précision des nombres et les ID sont des chaînes. L'adresse par défaut est https://api.requestbeam.com sans /v1. Une adresse personnalisée doit être une origine HTTPS de confiance ; les redirections sont bloquées.

`en`, `de`, `tr`, `it`, `es`, `fr`, `pt-BR`, `ja`, `ko`, `zh-CN`, `zh-TW`, `ru`.

## Streaming et tâches

StreamChatCompletions, StreamResponses, StreamMessages et StreamImage transmettent au callback des SSEEvent avec Event, Data et ID. Chaque événement ne contient pas forcément du texte. Une erreur du callback ou l'annulation du context arrête la lecture ; [DONE] est traité en interne. Les images renvoient directement du JSON ou du SSE, sans Jobs vidéo. GenerateVideo renvoie data.jobId. WaitForJob termine à COMPLETED ; FAILED et CANCELLED produisent JobFailedError. Le délai ou l'annulation arrête seulement l'attente locale, pas la tâche distante ni sa facturation.

Le délai par défaut de 60 secondes couvre toute la réponse, streaming compris. Pour une génération longue, utilisez `requestbeam.WithTimeout(10*time.Minute)` ; un délai plus court du context ou du client HTTP reste applicable. Chaque événement SSE est limité à 64 MiB. `response.incomplete` termine le flux mais indique un résultat incomplet : vérifiez le statut et `incomplete_details`.

[StreamChatCompletions](../examples/stream/main.go) · [GenerateVideo / WaitForJob](../examples/video/main.go)

## Similarweb et Twitter

Similarweb prend en charge 27 opérations et uniquement des données mondiales : country=999. Les autres pays acceptés donnent aussi des données mondiales. Fournissez from et to ensemble ; le SDK ne modifie ni les dates ni les champs natifs. Des résultats vides ne prouvent pas un trafic nul. TwitterUserByUsername, TwitterUser et TwitterUserTweets renvoient les réponses natives. TwitterUserTweetPages suit meta.next_token avec une limite positive de pages et rejette les curseurs répétés. Chaque page peut être facturée.

[Similarweb](../examples/data/main.go) · [API](api-reference.md)

## Erreurs et vérification

APIError conserve le statut HTTP, le code du service, l'ID de requête et la réponse originale. SDKError identifie les erreurs locales ; PollingTimeoutError et JobFailedError distinguent délai dépassé et échec de tâche. Le SDK lui-même n'implémente pas de nouvelles tentatives automatiques, même pour GET. Une erreur réseau n'exclut pas un traitement ou une facturation. Gardez les clés hors de Git et des journaux. Téléchargez les fichiers avec un client séparé sans authentification. Les tests utilisent des simulations locales ; les exemples envoient de vraies requêtes potentiellement payantes.

```bash
go test -race ./...
go vet ./...
gofmt -l .
```

[API](api-reference.md) · [Design](design.md) · [RequestBeam](https://requestbeam.com/fr)
