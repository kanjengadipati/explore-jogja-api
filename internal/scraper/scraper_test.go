package scraper

import (
	"testing"
	"time"
)

func TestParseInjourneyEventDates(t *testing.T) {
	cases := []struct {
		title, content, published string
		wantStart, wantEnd        string
	}{
		{"Waisak Borobudur 2570 BE", "", "", "2027-01-01", "2027-12-31"},
		{"Prambanan Jazz 2018", "", "", "2018-01-01", "2018-12-31"},
		{"YANNI Live Orchestra at Prambanan Temple", "<p>October 20th, 2018</p>", "", "2018-10-20", "2018-10-20"},
		{"Sendratari Ramayana Prambanan", "<p>20 Oktober 2018</p>", "", "2018-10-20", "2018-10-20"},
		{"Test Event", "<p>2018-10-20</p>", "", "2018-10-20", "2018-10-20"},
		{"Test Event", "<p>20/10/2018</p>", "", "2018-10-20", "2018-10-20"},
		{"Test Event", "", "2024-11-07T14:30:57", "2024-11-07", "2024-11-07"},
		{"Test Event", "", "", "", ""},
	}
	for _, c := range cases {
		s, e := parseInjourneyEventDates(c.title, c.content, c.published)
		if s != c.wantStart || e != c.wantEnd {
			t.Errorf("parseInjourneyEventDates(%q,%q,%q) = (%q,%q), want (%q,%q)", c.title, c.content, c.published, s, e, c.wantStart, c.wantEnd)
		}
	}
}

func TestMatchFromDestMap(t *testing.T) {
	m := map[string]string{
		"ratu-boko":  "ratu boko",
		"boko":       "boko",
		"borobudur":  "candi borobudur",
		"prambanan":  "prambanan",
		"jogjakarta": "yogyakarta",
		"jogja":      "jogja",
	}
	cases := []struct {
		title, location string
		want            string
	}{
		{"Festival Ratu Boko 2024", "", "ratu-boko"},
		{"Sendratari Ramayana Prambanan", "", "prambanan"},
		{"Virtual Wellness Tour Candi Borobudur", "", "borobudur"},
		{"Jogjakarta Half Marathon", "", ""},
		{"Yogya Run", "Yogyakarta", "jogjakarta"},
		{"Nothing matches", "", ""},
	}
	for _, c := range cases {
		got := matchFromDestMap(m, c.title, c.location)
		if got != c.want {
			t.Errorf("matchFromDestMap(%q,%q) = %q, want %q", c.title, c.location, got, c.want)
		}
	}
}

func TestIsStaleEvent(t *testing.T) {
	longAgo := time.Now().AddDate(0, 0, -40).Format("2006-01-02")
	recent := time.Now().AddDate(0, 0, -5).Format("2006-01-02")

	cases := []struct {
		name      string
		item      ScrapedEvent
		wantStale bool
	}{
		{"old end date", ScrapedEvent{StartDate: longAgo, EndDate: longAgo}, true},
		{"recent end date", ScrapedEvent{StartDate: recent, EndDate: recent}, false},
		{"old start, no end", ScrapedEvent{StartDate: longAgo}, true},
		{"no dates", ScrapedEvent{}, false},
		{"unparseable date", ScrapedEvent{EndDate: "not-a-date"}, false},
	}
	for _, c := range cases {
		if got := isStaleEvent(c.item); got != c.wantStale {
			t.Errorf("%s: isStaleEvent(%+v) = %v, want %v", c.name, c.item, got, c.wantStale)
		}
	}
}

func TestParseVisitingJogjaTitleDate(t *testing.T) {
	cases := []struct {
		title     string
		wantStart string
		wantEnd   string
	}{
		{"Menoreh Harmony Festival 2026 (8 Agustus 2026)", "2026-08-08", "2026-08-08"},
		{"DJOGJANTIQUEDAY 10 GASSAWARSA (7–8 Agustus 2026)", "2026-08-07", "2026-08-08"},
		{"Yogyakarta Gamelan Festival 2026 (21 Juli-2 Agustus 2026)", "2026-07-21", "2026-08-02"},
		{"Yogyakarta Gamelan Festival 2026 (21 Juli - 2 Agustus 2026)", "2026-07-21", "2026-08-02"},
		{"Jogja Fashion Week 2026 (13–16 August 2026)", "2026-08-13", "2026-08-16"},
		{"Konser Tanpa Judul (no date here)", "", ""},
		{"Sebuah Event Tanpa Kurung 2026", "", ""},
	}
	for _, c := range cases {
		start, end := parseVisitingJogjaTitleDate(c.title)
		if start != c.wantStart || end != c.wantEnd {
			t.Errorf("parseVisitingJogjaTitleDate(%q) = (%q, %q), want (%q, %q)",
				c.title, start, end, c.wantStart, c.wantEnd)
		}
	}
}
