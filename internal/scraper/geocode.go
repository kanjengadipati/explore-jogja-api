package scraper

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var geocodeCache = struct {
	sync.Mutex
	m map[string]geoResult
}{m: make(map[string]geoResult)}

// Nominatim's usage policy allows ~1 request per second. The ticker paces
// cache-miss lookups so a large scrape never bursts past the limit.
var geocodeLimiter = time.NewTicker(1100 * time.Millisecond)

// jogjaRegionViewbox bounds the area this platform publishes content for:
// DI Yogyakarta plus the immediately adjacent Magelang/Borobudur corridor
// (the injourney scraper deliberately includes Borobudur, which sits just
// outside DIY proper). Combined with bounded=1, Nominatim only returns
// results inside it, so an ambiguous location string can never resolve to
// coordinates in a far-away province.
const jogjaRegionViewbox = "109.85,-7.4,110.9,-8.3"

type geoResult struct {
	lat float64
	lng float64
	ok  bool
}

func geocode(location string) (lat, lng float64, ok bool) {
	loc := strings.TrimSpace(location)
	if loc == "" {
		return 0, 0, false
	}
	key := strings.ToLower(loc)

	geocodeCache.Lock()
	if r, found := geocodeCache.m[key]; found {
		geocodeCache.Unlock()
		return r.lat, r.lng, r.ok
	}
	geocodeCache.Unlock()

	<-geocodeLimiter.C

	reqURL := "https://nominatim.openstreetmap.org/search?format=json&limit=1&bounded=1&viewbox=" +
		jogjaRegionViewbox + "&q=" + url.QueryEscape(loc)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return 0, 0, false
	}
	req.Header.Set("User-Agent", "ExploreJogja/1.0 (jogja travel platform)")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, false
	}
	defer resp.Body.Close()

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil || len(results) == 0 {
		geocodeCache.Lock()
		geocodeCache.m[key] = geoResult{ok: false}
		geocodeCache.Unlock()
		return 0, 0, false
	}

	lat, err1 := strconv.ParseFloat(results[0].Lat, 64)
	lng, err2 := strconv.ParseFloat(results[0].Lon, 64)
	if err1 != nil || err2 != nil {
		geocodeCache.Lock()
		geocodeCache.m[key] = geoResult{ok: false}
		geocodeCache.Unlock()
		return 0, 0, false
	}

	geocodeCache.Lock()
	geocodeCache.m[key] = geoResult{lat: lat, lng: lng, ok: true}
	geocodeCache.Unlock()
	return lat, lng, true
}
