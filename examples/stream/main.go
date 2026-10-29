package main

import (
	"context"
	"fmt"
	"log"
	"os"

	requestbeam "github.com/RequestBeamDev/golang_sdk"
)

func main() {
	model := os.Getenv("REQUESTBEAM_TEXT_MODEL")
	if model == "" {
		log.Fatal("Set REQUESTBEAM_TEXT_MODEL to a model available to your account")
	}
	client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.StreamChatCompletions(context.Background(), map[string]any{
		"model":    model,
		"messages": []map[string]any{{"role": "user", "content": "Hello"}},
	}, func(event requestbeam.SSEEvent) error {
		fmt.Println(event.Event, event.Data)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
