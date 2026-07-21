package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"pleco-api/internal/config"
	"pleco-api/internal/modules/destination"
)

func main() {
	config.LoadEnv()
	appConfig := config.LoadAppConfig()
	db := config.ConnectDBWithDriver(appConfig.DatabaseURL, appConfig.DatabaseDriver)

	// Fetch all destinations with video_url from DB
	var dbDests []destination.Destination
	if err := db.Where("video_url IS NOT NULL AND video_url <> ''").Find(&dbDests).Error; err != nil {
		log.Fatalf("Failed to fetch destinations from DB: %v", err)
	}

	fmt.Printf("Fetched %d destinations with video_url from DB.\n", len(dbDests))
	if len(dbDests) == 0 {
		fmt.Println("No video URLs found in DB. Run the fetcher script first.")
		return
	}

	// Create a map of ExternalID -> VideoURL
	videoMap := make(map[string]string)
	for _, d := range dbDests {
		videoMap[d.ExternalID] = d.VideoURL
	}

	// Write to temporary JSON file
	tmpFile := "/tmp/videos.json"
	tmpData, err := json.Marshal(videoMap)
	if err != nil {
		log.Fatalf("Failed to marshal temporary video map: %v", err)
	}
	if err := os.WriteFile(tmpFile, tmpData, 0644); err != nil {
		log.Fatalf("Failed to write temporary video map: %v", err)
	}

	// Run python script to merge, preserving key order
	pyCode := `
import json

with open("/tmp/videos.json") as f:
    video_map = json.load(f)

json_path = "internal/seeds/destinations.json"
with open(json_path) as f:
    json_dests = json.load(f)

updated_count = 0
for d in json_dests:
    id_val = d.get("id")
    if id_val in video_map:
        d["video_url"] = video_map[id_val]
        updated_count += 1

with open(json_path, "w", encoding="utf-8") as f:
    json.dump(json_dests, f, indent=2, ensure_ascii=False)

print(f"Successfully updated destinations.json seed file: {updated_count} items updated with video_url.")
`
	cmd := exec.Command("python3", "-c", pyCode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Python script failed: %v", err)
	}

	// Cleanup
	_ = os.Remove(tmpFile)
}
