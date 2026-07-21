package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"pleco-api/internal/config"
	"pleco-api/internal/modules/destination"
)

type ytItem struct {
	ID struct {
		VideoID string `json:"videoId"`
	} `json:"id"`
	Snippet struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ChannelID   string `json:"channelId"`
		ChannelTitle string `json:"channelTitle"`
		PublishedAt string `json:"publishedAt"`
	} `json:"snippet"`
}

type ytResponse struct {
	Items []ytItem `json:"items"`
}

type ytVideoItem struct {
	ID      string `json:"id"`
	Snippet struct {
		ChannelTitle string `json:"channelTitle"`
	} `json:"snippet"`
	Statistics struct {
		ViewCount string `json:"viewCount"`
		LikeCount string `json:"likeCount"`
	} `json:"statistics"`
	ContentDetails struct {
		Duration string `json:"duration"`
		Definition string `json:"definition"`
	} `json:"contentDetails"`
}

type ytVideoResponse struct {
	Items []ytVideoItem `json:"items"`
}

func main() {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY env not set")
	}

	config.LoadEnv()
	appConfig := config.LoadAppConfig()
	db := config.ConnectDBWithDriver(appConfig.DatabaseURL, appConfig.DatabaseDriver)

	var dests []destination.Destination
	db.Where("video_url = '' OR video_url IS NULL").Find(&dests)
	log.Printf("Found %d destinations without video_url", len(dests))

	client := http.Client{Timeout: 15 * time.Second}

	updated := 0
	skipped := 0

	for i, d := range dests {
		videoID, err := searchBestVideo(client, apiKey, d.Name)
		if err != nil {
			log.Printf("[%d/%d] %s: %v", i+1, len(dests), d.Name, err)
			skipped++
			continue
		}
		if videoID == "" {
			log.Printf("[%d/%d] %s: no video found", i+1, len(dests), d.Name)
			skipped++
			continue
		}

		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
		db.Model(&d).Update("video_url", videoURL)
		log.Printf("[%d/%d] ✓ %s → %s", i+1, len(dests), d.Name, videoURL)
		updated++

		// rate limit: 1 request per 300ms to avoid quota burst
		time.Sleep(300 * time.Millisecond)

		// stop if we hit ~100 queries (quota: 10000/day, 1 search = 100 units)
		if updated >= 100 {
			log.Printf("Reached 100 updates for today. Run again tomorrow for the rest.")
			break
		}
	}

	log.Printf("Done. Updated: %d, Skipped: %d", updated, skipped)
}

func searchBestVideo(client http.Client, apiKey, query string) (string, error) {
	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("q", query+" Yogyakarta wisata")
	q.Set("type", "video")
	q.Set("maxResults", "5")
	q.Set("relevanceLanguage", "id")
	q.Set("videoDuration", "medium")
	q.Set("videoDefinition", "high")
	q.Set("key", apiKey)

	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search?"+q.Encode(), nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ytResp ytResponse
	if err := json.Unmarshal(body, &ytResp); err != nil {
		return "", fmt.Errorf("search parse: %w", err)
	}

	if len(ytResp.Items) == 0 {
		return "", nil
	}

	// collect video IDs for stats lookup
	var videoIDs []string
	for _, item := range ytResp.Items {
		videoIDs = append(videoIDs, item.ID.VideoID)
	}

	// get video stats to pick best quality
	best := pickBestVideo(client, apiKey, videoIDs)
	return best, nil
}

func pickBestVideo(client http.Client, apiKey string, videoIDs []string) string {
	if len(videoIDs) == 0 {
		return ""
	}

	q := url.Values{}
	q.Set("part", "snippet,contentDetails,statistics")
	q.Set("id", strings.Join(videoIDs, ","))
	q.Set("key", apiKey)

	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/videos?"+q.Encode(), nil)
	resp, err := client.Do(req)
	if err != nil {
		return videoIDs[0] // fallback
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var ytResp ytVideoResponse
	if err := json.Unmarshal(body, &ytResp); err != nil {
		return videoIDs[0]
	}

	if len(ytResp.Items) == 0 {
		return videoIDs[0]
	}

	// score: prefer HD, prefer more views
	type scored struct {
		id    string
		score int
	}
	var scoredVideos []scored
	for _, v := range ytResp.Items {
		s := 0
		if v.ContentDetails.Definition == "hd" {
			s += 100
		}
		views := 0
		fmt.Sscanf(v.Statistics.ViewCount, "%d", &views)
		s += views / 1000
		scoredVideos = append(scoredVideos, scored{id: v.ID, score: s})
	}

	sort.Slice(scoredVideos, func(i, j int) bool {
		return scoredVideos[i].score > scoredVideos[j].score
	})

	return scoredVideos[0].id
}
