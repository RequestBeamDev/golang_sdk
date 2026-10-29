package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	requestbeam "github.com/RequestBeamDev/golang_sdk"
)

func main() {
	model := os.Getenv("REQUESTBEAM_TEXT_MODEL")
	if model == "" {
		log.Fatal("Set REQUESTBEAM_TEXT_MODEL to a model available to your account")
	}
	client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"), requestbeam.WithLocale("en"))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.ChatCompletions(context.Background(), map[string]any{
		"model":    model,
		"messages": []map[string]any{{"role": "user", "content": "Hello"}},
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}
