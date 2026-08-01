package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	injourneyBase = "https://injourneydestination.id"
	// WordPress REST API endpoints
	injourneyDestAPI  = injourneyBase + "/wp-json/wp/v2/experience?per_page=100"
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
	ID    int    `json:"id"`
	Slug  string `json:"slug"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Excerpt struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
	FeaturedMedia      int   `json:"featured_media"`
	ExperienceCategory []int `json:"experience_category"`
}

type wpCategoryResponse struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
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

// injourneyCategoryKeywords classifies an "experience" post into a real
// destination category based on keywords in its title, instead of dumping
// every post into a single generic "Cultural" bucket. Checked in order;
// first match wins. Anything unmatched falls back to "heritage" since every
// injourney experience is an add-on tied to a heritage site (Borobudur,
// Prambanan, Ratu Boko).
var injourneyCategoryKeywords = []struct {
	category string
	keywords []string
}{
	{"culinary", []string{"barbekyu", "bbq", "meals", "dhaharan", "racik", "picnic"}},
	{"camping", []string{"camping"}},
	{"adventure", []string{"cycling", "trekking", "outbound"}},
	{"family", []string{"sinema", "cinema"}},
}

func classifyInjourneyCategory(title string) string {
	lower := strings.ToLower(title)
	for _, rule := range injourneyCategoryKeywords {
		for _, kw := range rule.keywords {
			if strings.Contains(lower, kw) {
				return rule.category
			}
		}
	}
	return "heritage"
}

// fetchCategories loads the "experience_category" taxonomy. The InJourney
// experience/event posts are add-on packages tied to a specific heritage site
// (Borobudur, Prambanan, Ratu Boko) and the taxonomy tells us which one — this
// is the only reliable source of location, so we never hardcode "Yogyakarta".
func (s *injourneyScraper) fetchCategories() (map[int]string, error) {
	url := injourneyBase + "/wp-json/wp/v2/experience_category?per_page=100"
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

	var cats []wpCategoryResponse
	if err := json.Unmarshal(body, &cats); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	m := make(map[int]string, len(cats))
	for _, c := range cats {
		m[c.ID] = c.Slug
	}
	return m, nil
}

// injourneySiteLocations maps the InJourney site taxonomy slugs to real
// locations/sub-regions. Borobudur sits in Magelang (Central Java), Prambanan
// and Ratu Boko are in Sleman (DIY).
func injourneyLocationFor(slug string) (location, subRegion string) {
	switch slug {
	case "borobudur":
		return "Borobudur, Magelang", "Magelang"
	case "prambanan", "ramayana":
		return "Prambanan, Sleman", "Sleman"
	case "ratu-boko":
		return "Ratu Boko, Sleman", "Sleman"
	}
	return "", ""
}

func injourneyFallbackLocation() (string, string) {
	return "Yogyakarta", "Yogyakarta"
}

func injourneyEventLocation(title string) (location, subRegion string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "borobudur"):
		return "Borobudur, Magelang", "Magelang"
	case strings.Contains(lower, "ratu boko"):
		return "Ratu Boko, Sleman", "Sleman"
	case strings.Contains(lower, "prambanan") || strings.Contains(lower, "ramayana"):
		return "Prambanan, Sleman", "Sleman"
	}
	return injourneyFallbackLocation()
}

var (
	injourneyBEPattern = regexp.MustCompile(`\b(\d{4})\s*BE\b`)
	injourneyCEPattern = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
)

// parseInjourneyEventDates extracts a best-effort year from the event title
// (e.g. "Prambanan Jazz 2018" → 2018, "Waisak Borobudur 2570 BE" → 2027 CE) so
// event statuses (completed/active/upcoming) are meaningful. Returns empty
// strings when no year can be found.
func parseInjourneyEventDates(title string) (start, end string) {
	if m := injourneyBEPattern.FindStringSubmatch(title); m != nil {
		year := 0
		fmt.Sscanf(m[1], "%d", &year)
		year -= 543 // Buddhist Era → CE
		if year >= 1990 && year <= 2100 {
			return fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
		}
	}
	if m := injourneyCEPattern.FindStringSubmatch(title); m != nil {
		return m[1] + "-01-01", m[1] + "-12-31"
	}
	return "", ""
}

func (s *injourneyScraper) ScrapeDestinations() ([]ScrapedDestination, error) {
	posts, err := s.fetchJSON(injourneyDestAPI)
	if err != nil {
		return nil, fmt.Errorf("fetch injourney destinations: %w", err)
	}

	cats, _ := s.fetchCategories()

	var dests []ScrapedDestination
	for _, p := range posts {
		title := strings.TrimSpace(p.Title.Rendered)
		if title == "" {
			continue
		}

		location, subRegion := "", ""
		for _, id := range p.ExperienceCategory {
			if loc, sub := injourneyLocationFor(cats[id]); sub != "" {
				location, subRegion = loc, sub
				break
			}
		}
		if subRegion == "" {
			location, subRegion = injourneyEventLocation(title)
		}

		img := s.fetchMedia(p.FeaturedMedia)
		desc := stripHTML(p.Excerpt.Rendered)

		dests = append(dests, ScrapedDestination{
			ExternalID:  slugify(title),
			Name:        title,
			Tagline:     "",
			Category:    classifyInjourneyCategory(title),
			Location:    location,
			SubRegion:   subRegion,
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
		location, _ := injourneyEventLocation(title)
		startDate, endDate := parseInjourneyEventDates(title)

		events = append(events, ScrapedEvent{
			ExternalID:  slugify(title),
			Title:       title,
			Description: desc,
			Location:    location,
			StartDate:   startDate,
			EndDate:     endDate,
			ImageURL:    img,
			Category:    "Event",
			TicketPrice: "Cek website resmi",
			Organizer:   "InJourney Destination Management",
			Source:      "injourney",
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
