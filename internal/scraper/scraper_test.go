package scraper

import "testing"

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
