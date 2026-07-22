package scraper

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	jadestaBase   = "https://jadesta.kemenpar.go.id"
	jadestaSearch = jadestaBase + "/desa?search=yogyakarta&prov=34"
)

type jadestaScraper struct {
	client http.Client
}

func init() {
	Register(&jadestaScraper{
		client: http.Client{Timeout: 15 * time.Second},
	})
}

func (s *jadestaScraper) Name() string {
	return "jadesta"
}

func (s *jadestaScraper) fetchDoc(url string) (*goquery.Document, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ExploreJogja/1.0)")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromReader(strings.NewReader(string(body)))
}

func (s *jadestaScraper) ScrapeEvents() ([]ScrapedEvent, error) {
	// Jadesta only lists tourism villages (desa wisata), not events.
	return nil, nil
}

func (s *jadestaScraper) ScrapeDestinations() ([]ScrapedDestination, error) {
	doc, err := s.fetchDoc(jadestaSearch)
	if err != nil {
		return nil, fmt.Errorf("fetch jadesta destinations: %w", err)
	}

	var dests []ScrapedDestination
	doc.Find(".listing-item-content").Each(func(i int, sel *goquery.Selection) {
		name := strings.TrimSpace(sel.Find("h3").First().Text())
		if name == "" {
			return
		}

		location := "Yogyakarta"
		subRegion := "Yogyakarta"
		spanText := strings.TrimSpace(sel.Find("span").First().Text())
		if spanText != "" {
			parts := strings.SplitN(spanText, ", ", 2)
			if len(parts) == 2 {
				location = strings.TrimSpace(parts[0])
				subRegion = normalizeSubRegion(parts[len(parts)-1])
			} else {
				subRegion = normalizeSubRegion(spanText)
			}
		}

		dests = append(dests, ScrapedDestination{
			ExternalID:  slugify(name),
			Name:        name,
			Tagline:     "",
			Category:    "Desa Wisata",
			Location:    location,
			SubRegion:   subRegion,
			Images:      nil,
			Description: "",
			TicketPrice: "Cek website resmi",
			Source:      "jadesta",
		})
	})

	return dests, nil
}
