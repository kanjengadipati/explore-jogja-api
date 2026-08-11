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
	visitingJogjaBase = "https://visitingjogja.jogjaprov.go.id"
	// WordPress REST API. Category 4 = "Event" (individual event posts).
	// Category 356 = "Calendar of Event" (monthly roundup posts that bundle
	// many events into one article) and must be excluded — it is not a
	// single event and has no reliable single date/location.
	visitingJogjaEventCategory      = 4
	visitingJogjaCalendarRoundupCat = 356
	visitingJogjaEventAPI           = visitingJogjaBase + "/wp-json/wp/v2/posts?categories=%d&per_page=100&page=%d"
)

type visitingJogjaScraper struct {
	client http.Client
}

func init() {
	Register(&visitingJogjaScraper{
		client: http.Client{Timeout: 30 * time.Second},
	})
}

func (s *visitingJogjaScraper) Name() string {
	return "visitingjogja"
}

func (s *visitingJogjaScraper) ScrapeDestinations() ([]ScrapedDestination, error) {
	// This source only publishes events, not destinations.
	return nil, nil
}

type vjMediaResponse struct {
	SourceURL string `json:"source_url"`
}

type vjPostResponse struct {
	ID    int    `json:"id"`
	Slug  string `json:"slug"`
	Link  string `json:"link"`
	Date  string `json:"date"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
	Excerpt struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
	FeaturedMedia int   `json:"featured_media"`
	Categories    []int `json:"categories"`
}

func (s *visitingJogjaScraper) fetchPage(page int) ([]vjPostResponse, error) {
	url := fmt.Sprintf(visitingJogjaEventAPI, visitingJogjaEventCategory, page)
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

	// WordPress returns 400 once `page` exceeds the total page count —
	// treat that as "no more pages" rather than an error.
	if resp.StatusCode == http.StatusBadRequest {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var posts []vjPostResponse
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}
	return posts, nil
}

func (s *visitingJogjaScraper) fetchMedia(id int) string {
	if id == 0 {
		return ""
	}
	url := fmt.Sprintf("%s/wp-json/wp/v2/media/%d", visitingJogjaBase, id)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ExploreJogja/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var media vjMediaResponse
	if err := json.NewDecoder(resp.Body).Decode(&media); err != nil {
		return ""
	}
	return media.SourceURL
}

func hasCategory(cats []int, id int) bool {
	for _, c := range cats {
		if c == id {
			return true
		}
	}
	return false
}

// --- Title date parsing -----------------------------------------------
//
// Every individual event post title ends with a parenthesized date range,
// e.g. "Menoreh Harmony Festival 2026 (8 Agustus 2026)" or
// "Jogja Fashion Week 2026 (13–16 August 2026)" or
// "Yogyakarta Gamelan Festival 2026 (21 Juli-2 Agustus 2026 )".
// This is far more reliable than scanning free-text body content.

var (
	vjMonthNumber = map[string]int{
		"january": 1, "januari": 1, "jan": 1,
		"february": 2, "februari": 2, "feb": 2,
		"march": 3, "maret": 3, "mar": 3,
		"april": 4, "apr": 4,
		"may": 5, "mei": 5,
		"june": 6, "juni": 6, "jun": 6,
		"july": 7, "juli": 7, "jul": 7,
		"august": 8, "agustus": 8, "aug": 8, "ags": 8,
		"september": 9, "sep": 9, "sept": 9,
		"october": 10, "oktober": 10, "oct": 10, "okt": 10,
		"november": 11, "nov": 11,
		"december": 12, "desember": 12, "dec": 12, "des": 12,
	}

	// Matches the parenthesized tail of the title. Captures everything
	// between the last "(" and the final ")".
	vjParenPattern = regexp.MustCompile(`\(([^()]+)\)\s*$`)

	// "8 Agustus 2026" or "8 August 2026"
	vjSingleDatePattern = regexp.MustCompile(`(?i)^(\d{1,2})\s+([a-z]+)\s+(\d{4})$`)

	// "7–8 Agustus 2026" / "7-8 Agustus 2026" / "13–16 August 2026"
	// (same month, day range)
	vjSameMonthRangePattern = regexp.MustCompile(`(?i)^(\d{1,2})\s*[-–]\s*(\d{1,2})\s+([a-z]+)\s+(\d{4})$`)

	// "21 Juli-2 Agustus 2026" (day+month on both ends, spanning months)
	vjCrossMonthRangePattern = regexp.MustCompile(`(?i)^(\d{1,2})\s+([a-z]+)\s*[-–]\s*(\d{1,2})\s+([a-z]+)\s+(\d{4})$`)
)

// parseVisitingJogjaTitleDate extracts start/end (YYYY-MM-DD) from the
// parenthesized date range at the end of an event title. Returns "","" if no
// recognizable pattern is found — callers should fall back to the publish
// date so the event still gets a sane status instead of being "upcoming"
// forever.
func parseVisitingJogjaTitleDate(title string) (start, end string) {
	m := vjParenPattern.FindStringSubmatch(title)
	if m == nil {
		return "", ""
	}
	raw := strings.TrimSpace(m[1])

	if mm := vjCrossMonthRangePattern.FindStringSubmatch(raw); mm != nil {
		startDay, _ := strconv.Atoi(mm[1])
		startMonth := vjMonthNumber[strings.ToLower(mm[2])]
		endDay, _ := strconv.Atoi(mm[3])
		endMonth := vjMonthNumber[strings.ToLower(mm[4])]
		year, _ := strconv.Atoi(mm[5])
		s := dateFromPartsInt(year, startMonth, startDay)
		e := dateFromPartsInt(year, endMonth, endDay)
		if s != "" && e != "" {
			return s, e
		}
	}

	if mm := vjSameMonthRangePattern.FindStringSubmatch(raw); mm != nil {
		startDay, _ := strconv.Atoi(mm[1])
		endDay, _ := strconv.Atoi(mm[2])
		month := vjMonthNumber[strings.ToLower(mm[3])]
		year, _ := strconv.Atoi(mm[4])
		s := dateFromPartsInt(year, month, startDay)
		e := dateFromPartsInt(year, month, endDay)
		if s != "" && e != "" {
			return s, e
		}
	}

	if mm := vjSingleDatePattern.FindStringSubmatch(raw); mm != nil {
		day, _ := strconv.Atoi(mm[1])
		month := vjMonthNumber[strings.ToLower(mm[2])]
		year, _ := strconv.Atoi(mm[3])
		d := dateFromPartsInt(year, month, day)
		if d != "" {
			return d, d
		}
	}

	return "", ""
}

// --- Body field extraction ------------------------------------------------
//
// Post bodies are short, emoji-led fact sheets rather than free prose, e.g.:
//   📍 Jogja Expo Center<br />
//   📅 7–8 Agustus 2026<br />
//   🎟️ Gratis untuk umum<br />
// or:
//   ✅ Jersey<br />
//   ✅ BIB Number<br />
//   ✅ Medali<br />
// vjPlainLines turns the HTML body into one plain-text line per <br>/<p>
// break so each fact can be pulled out by its leading emoji marker.

// Matches <br>, <br/>, <br />, <p>, <p class="...">, </p>, including variants
// with extra attributes injected by the CMS's rich-text editor (e.g.
// `<br data-start="617" data-end="620" />`) that a fixed string replace would miss.
var vjBlockTagPattern = regexp.MustCompile(`(?i)</?(?:br|p)(?:\s[^>]*)?/?>`)

func vjPlainLines(html string) []string {
	text := stripHTML(vjBlockTagPattern.ReplaceAllString(html, "\n"))
	raw := strings.Split(text, "\n")
	lines := make([]string, 0, len(raw))
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// vjExtractAfterMarker returns the trimmed text following the first line
// that contains one of the given emoji markers (checked in order), e.g.
// markers=["📍"] on "📍 Jogja Expo Center" → "Jogja Expo Center".
func vjExtractAfterMarker(lines []string, markers []string) string {
	for _, line := range lines {
		for _, m := range markers {
			if idx := strings.Index(line, m); idx != -1 {
				rest := strings.TrimSpace(line[idx+len(m):])
				if rest != "" {
					return rest
				}
			}
		}
	}
	return ""
}

// vjExtractHighlights collects short bullet-style lines that start with one
// of the given markers (✨/✅ are used for feature/benefit lists such as
// "✅ Jersey", "✨ 150+ Desainer Fashion"). A line ending in ":" is treated
// as a section heading (e.g. "✨ Penampilan spesial dari:") and skipped, not
// as a highlight itself.
func vjExtractHighlights(lines []string, markers []string, max int) []string {
	var out []string
	for _, line := range lines {
		for _, m := range markers {
			if !strings.HasPrefix(line, m) {
				continue
			}
			text := strings.TrimSpace(strings.TrimPrefix(line, m))
			if text == "" || strings.HasSuffix(text, ":") {
				break
			}
			out = append(out, text)
			if len(out) >= max {
				return out
			}
			break
		}
	}
	return out
}

var (
	vjLocationMarkers  = []string{"📍"}
	vjTicketMarkers    = []string{"💵", "💰", "🎟️", "🎟", "🎫"}
	vjHighlightMarkers = []string{"✨", "✅"}
)

func (s *visitingJogjaScraper) ScrapeEvents() ([]ScrapedEvent, error) {
	var all []vjPostResponse
	for page := 1; page <= 10; page++ { // hard cap to avoid runaway pagination
		posts, err := s.fetchPage(page)
		if err != nil {
			return nil, fmt.Errorf("fetch visitingjogja events page %d: %w", page, err)
		}
		if len(posts) == 0 {
			break
		}
		all = append(all, posts...)
	}

	var events []ScrapedEvent
	for _, p := range all {
		// Skip monthly "Calendar of Event" roundup posts — not a single event.
		if hasCategory(p.Categories, visitingJogjaCalendarRoundupCat) {
			continue
		}

		title := strings.TrimSpace(stripHTML(p.Title.Rendered))
		if title == "" {
			continue
		}

		content := p.Content.Rendered
		lines := vjPlainLines(content)
		location := vjExtractAfterMarker(lines, vjLocationMarkers)
		ticketPrice := vjExtractAfterMarker(lines, vjTicketMarkers)
		highlights := vjExtractHighlights(lines, vjHighlightMarkers, 10)
		desc := postDescription(p.Excerpt.Rendered, p.Content.Rendered)

		startDate, endDate := parseVisitingJogjaTitleDate(title)
		if startDate == "" && len(p.Date) >= 10 {
			// No parseable date in the title — fall back to publish date
			// so status resolution doesn't default to "upcoming" forever.
			if _, err := time.Parse("2006-01-02", p.Date[:10]); err == nil {
				startDate, endDate = p.Date[:10], p.Date[:10]
			}
		}

		img := s.fetchMedia(p.FeaturedMedia)

		events = append(events, ScrapedEvent{
			ExternalID:  slugify(title),
			Title:       title,
			Description: desc,
			Location:    location,
			StartDate:   startDate,
			EndDate:     endDate,
			ImageURL:    img,
			Category:    "Event",
			TicketPrice: ticketPrice,
			Organizer:   "Dinas Pariwisata DIY",
			Highlights:  highlights,
			Source:      "visitingjogja",
		})
	}

	return events, nil
}
