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
	Title      string `json:"title"`
	Statistics struct {
		ViewCount string `json:"viewCount"`
	} `json:"statistics"`
	ContentDetails struct {
		Definition string `json:"definition"`
	} `json:"contentDetails"`
}

var compilationKeywords = []string{
	"compilation", "kompilasi", "15 tempat", "10 tempat", "6 wisata",
	"8 wisata", "5 wisata", "tempat wisata di gunung", "tempat wisata di jogja",
	"wisata jogja", "jogja terbaru", "lagi hits", "paling terkenal",
	"wajib kunjung", "rekomendasi", "top 10", "top 5", "daftar",
}

func isCompilationTitle(title string) bool {
	lower := strings.ToLower(title)
	for _, kw := range compilationKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func FetchYouTubeVideoURL(query string) string {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return ""
	}

	client := http.Client{Timeout: 10 * time.Second}
	
	// 1. Search — use specific query without generic "wisata" suffix
	q := url.Values{}
	q.Set("part", "snippet")
	q.Set("q", query)
	q.Set("type", "video")
	q.Set("order", "relevance")
	q.Set("maxResults", "10")
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
	q.Set("part", "contentDetails,statistics,snippet")
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
	var candidates []scored
	for _, v := range videoResp.Items {
		if isCompilationTitle(v.Title) {
			continue
		}
		s := 0
		if v.ContentDetails.Definition == "hd" { s += 100 }
		views := 0
		fmt.Sscanf(v.Statistics.ViewCount, "%d", &views)
		s += views / 10000
		candidates = append(candidates, scored{v.ID, s})
	}

	if len(candidates) == 0 {
		return videoIDs[0]
	}

	var best scored
	for _, c := range candidates {
		if c.score > best.score {
			best = c
		}
	}
	return best.id
}
