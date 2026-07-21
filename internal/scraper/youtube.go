package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type ytItem struct {
	ID struct {
		VideoID string `json:"videoId"`
	} `json:"id"`
}

type ytVideoItem struct {
	ID         string `json:"id"`
	Statistics struct {
		ViewCount string `json:"viewCount"`
	} `json:"statistics"`
	ContentDetails struct {
		Definition string `json:"definition"`
	} `json:"contentDetails"`
}

func FetchYouTubeVideoURL(query string) string {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return ""
	}

	client := http.Client{Timeout: 10 * time.Second}
	
	// 1. Search
	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("q", query+" Yogyakarta wisata")
	q.Set("type", "video")
	q.Set("maxResults", "3")
	q.Set("relevanceLanguage", "id")
	q.Set("key", apiKey)

	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search?"+q.Encode(), nil)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var searchResp struct{ Items []ytItem }
	json.Unmarshal(body, &searchResp)

	if len(searchResp.Items) == 0 {
		return ""
	}

	// 2. Pick best
	videoIDs := make([]string, len(searchResp.Items))
	for i, item := range searchResp.Items {
		videoIDs[i] = item.ID.VideoID
	}
	
	videoID := pickBestVideo(client, apiKey, videoIDs)
	if videoID == "" {
		return ""
	}
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}

func pickBestVideo(client http.Client, apiKey string, videoIDs []string) string {
	q := url.Values{}
	q.Set("part", "contentDetails,statistics")
	q.Set("id", strings.Join(videoIDs, ","))
	q.Set("key", apiKey)

	req, _ := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/videos?"+q.Encode(), nil)
	resp, err := client.Do(req)
	if err != nil {
		return videoIDs[0]
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var videoResp struct{ Items []ytVideoItem }
	json.Unmarshal(body, &videoResp)

	if len(videoResp.Items) == 0 {
		return videoIDs[0]
	}

	type scored struct { id string; score int }
	var best scored
	for _, v := range videoResp.Items {
		s := 0
		if v.ContentDetails.Definition == "hd" { s += 100 }
		views := 0
		fmt.Sscanf(v.Statistics.ViewCount, "%d", &views)
		s += views / 1000
		if s > best.score {
			best = scored{v.ID, s}
		}
	}
	return best.id
}
