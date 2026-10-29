package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"

	requestbeam "github.com/RequestBeamDev/golang_sdk"
)

func main() {
	client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	result, err := client.Similarweb(context.Background(), "websites/visits", url.Values{
		"keys": {"example.com"}, "country": {"999"},
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}
