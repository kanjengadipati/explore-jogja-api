package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
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
	Date  string `json:"date"`
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

	// Concrete date patterns found in the post body ("October 20th, 2018",
	// "20 Oktober 2018", "2018-10-20"). Month names are English or Indonesian.
	injourneyMonthDayYearPattern = regexp.MustCompile(`(?i)\b(january|february|march|april|may|june|july|august|september|october|november|december|januari|februari|maret|april|mei|juni|juli|agustus|september|oktober|november|desember)[a-z]*\s+(\d{1,2})(?:st|nd|rd|th)?[,\s]+(\d{4})\b`)
	injourneyDayMonthYearPattern = regexp.MustCompile(`(?i)\b(\d{1,2})(?:st|nd|rd|th)?\s+(january|february|march|april|may|june|july|august|september|october|november|december|januari|februari|maret|april|mei|juni|juli|agustus|september|oktober|november|desember)[a-z]*[,\s]+(\d{4})\b`)
	injourneyNumericDatePattern  = regexp.MustCompile(`\b(\d{4})[-/.](\d{1,2})[-/.](\d{1,2})\b`)
	injourneyDMYDatePattern      = regexp.MustCompile(`\b(\d{1,2})[-/.](\d{1,2})[-/.](\d{4})\b`)
)

var injourneyMonthNumber = map[string]int{
	"january": 1, "januari": 1,
	"february": 2, "februari": 2,
	"march": 3, "maret": 3,
	"april": 4,
	"may":   5, "mei": 5,
	"june": 6, "juni": 6,
	"july": 7, "juli": 7,
	"august": 8, "agustus": 8,
	"september": 9,
	"october":   10, "oktober": 10,
	"november": 11,
	"december": 12, "desember": 12,
}

// parseInjourneyEventDates extracts the most specific date it can for an
// event post, in order of reliability:
//  1. a concrete date in the post body ("October 20th, 2018")
//  2. a year in the title ("Prambanan Jazz 2018", "Waisak Borobudur 2570 BE")
//  3. the post's publish date, so events without any explicit date still get
//     a meaningful status instead of being "upcoming" forever.
func parseInjourneyEventDates(title, content, published string) (start, end string) {
	if s := parseContentDate(content); s != "" {
		return s, s
	}
	if m := injourneyBEPattern.FindStringSubmatch(title); m != nil {
		year := 0
		fmt.Sscanf(m[1], "%d", &year)
		year -= 543 // Buddhist Era → CE
		if year >= 1990 && year <= 2100 {
			return fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
		}
	}
	if m := injourneyCEPattern.FindStringSubmatch(title); m != nil {
		if year, err := strconv.Atoi(m[1]); err == nil && year >= 1990 && year <= 2100 {
			return fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
		}
	}
	if len(published) >= 10 {
		if _, err := time.Parse("2006-01-02", published[:10]); err == nil {
			return published[:10], published[:10]
		}
	}
	return "", ""
}

// parseContentDate looks for a concrete date (month + day + year in either
// language ordering, or a numeric format) inside the post body.
func parseContentDate(content string) string {
	text := stripHTML(content)

	if m := injourneyMonthDayYearPattern.FindStringSubmatch(text); m != nil {
		if d := dateFromParts(m[3], m[2], injourneyMonthNumber[strings.ToLower(m[1])]); d != "" {
			return d
		}
	}
	if m := injourneyDayMonthYearPattern.FindStringSubmatch(text); m != nil {
		if d := dateFromParts(m[3], m[1], injourneyMonthNumber[strings.ToLower(m[2])]); d != "" {
			return d
		}
	}
	if m := injourneyNumericDatePattern.FindStringSubmatch(text); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		if d := dateFromPartsInt(year, month, day); d != "" {
			return d
		}
	}
	if m := injourneyDMYDatePattern.FindStringSubmatch(text); m != nil {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		if d := dateFromPartsInt(year, month, day); d != "" {
			return d
		}
	}
	return ""
}

// dateFromParts validates a day/month/year given as strings (month is a 1-12
// number) and formats it as YYYY-MM-DD, or "" when invalid.
func dateFromParts(yearStr, dayStr string, month int) string {
	year, err1 := strconv.Atoi(yearStr)
	day, err2 := strconv.Atoi(dayStr)
	if err1 != nil || err2 != nil {
		return ""
	}
	return dateFromPartsInt(year, month, day)
}

func dateFromPartsInt(year, month, day int) string {
	if year < 1990 || year > 2100 || month < 1 || month > 12 || day < 1 || day > 31 {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
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
		startDate, endDate := parseInjourneyEventDates(title, p.Content.Rendered, p.Date)

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
