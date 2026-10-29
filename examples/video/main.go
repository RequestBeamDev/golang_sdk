package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	requestbeam "github.com/RequestBeamDev/golang_sdk"
)

func main() {
	model := os.Getenv("REQUESTBEAM_VIDEO_MODEL")
	if model == "" {
		log.Fatal("Set REQUESTBEAM_VIDEO_MODEL to an available Seedance model")
	}
	client, err := requestbeam.NewClient(os.Getenv("REQUESTBEAM_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	submission, err := client.GenerateVideo(ctx, map[string]any{
		"model": model, "prompt": "A quiet city street at dawn, slow forward camera movement.",
		"duration": 5, "resolution": "720p", "aspectRatio": "16:9", "generateAudio": false,
	})
	if err != nil {
		log.Fatal(err)
	}
	data, ok := submission["data"].(map[string]any)
	if !ok {
		log.Fatal("Missing video submission data")
	}
	jobID, ok := data["jobId"].(string)
	if !ok || jobID == "" {
		log.Fatal("Missing string job ID")
	}
	job, err := client.WaitForJob(ctx, jobID, 10*time.Minute, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(job); err != nil {
		log.Fatal(err)
	}
}
