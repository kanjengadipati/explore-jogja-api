package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	injourneyBase = "https://injourneydestination.id"
	// WordPress REST API endpoints
	injourneyDestAPI = injourneyBase + "/wp-json/wp/v2/experience?per_page=100"
	injourneyEventAPI = injourneyBase + "/wp-json/wp/v2/event?per_page=100"
)

type injourneyScraper struct {
	client http.Client
}

func init() {
	Register(&injourneyScraper{
		client: http.Client{Timeout: 30 * time.Second},
	})
}

func (s *injourneyScraper) Name() string {
	return "injourney"
}

type wpMediaResponse struct {
	SourceURL string `json:"source_url"`
	AltText   string `json:"alt_text"`
}

type wpPostResponse struct {
	ID     int    `json:"id"`
	Slug   string `json:"slug"`
	Title  struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Excerpt struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
	FeaturedMedia int `json:"featured_media"`
}

func (s *injourneyScraper) fetchJSON(url string) ([]wpPostResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ExploreJogja/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var posts []wpPostResponse
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return posts, nil
}

func (s *injourneyScraper) fetchMedia(id int) string {
	if id == 0 {
		return ""
	}
	url := fmt.Sprintf("%s/wp-json/wp/v2/media/%d", injourneyBase, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ExploreJogja/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var media wpMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&media); err != nil {
		return ""
	}
	return media.SourceURL
}

func (s *injourneyScraper) ScrapeDestinations() ([]ScrapedDestination, error) {
	posts, err := s.fetchJSON(injourneyDestAPI)
	if err != nil {
		return nil, fmt.Errorf("fetch injourney destinations: %w", err)
	}

	var dests []ScrapedDestination
	for _, p := range posts {
		title := strings.TrimSpace(p.Title.Rendered)
		if title == "" {
			continue
		}

		img := s.fetchMedia(p.FeaturedMedia)
		desc := stripHTML(p.Excerpt.Rendered)

		dests = append(dests, ScrapedDestination{
			ExternalID:  slugify(title),
			Name:        title,
			Tagline:     "",
			Category:    "Cultural",
			Location:    "Yogyakarta",
			SubRegion:   "",
			Images:      imgs(img),
			Description: desc,
			TicketPrice: "Cek website resmi",
			Source:      "injourney",
		})
	}

	return dests, nil
}

func (s *injourneyScraper) ScrapeEvents() ([]ScrapedEvent, error) {
	posts, err := s.fetchJSON(injourneyEventAPI)
	if err != nil {
		return nil, fmt.Errorf("fetch injourney events: %w", err)
	}

	var events []ScrapedEvent
	for _, p := range posts {
		title := strings.TrimSpace(p.Title.Rendered)
		if title == "" {
			continue
		}

		img := s.fetchMedia(p.FeaturedMedia)
		desc := stripHTML(p.Excerpt.Rendered)

		events = append(events, ScrapedEvent{
			ExternalID:   slugify(title),
			Title:        title,
			Description:  desc,
			Location:     "Yogyakarta",
			ImageURL:     img,
			Category:     "Event",
			TicketPrice:  "Cek website resmi",
			Organizer:    "InJourney Destination Management",
			Source:       "injourney",
		})
	}

	return events, nil
}

func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			if r == '&' {
				result.WriteRune(' ')
			} else {
				result.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(result.String())
}
