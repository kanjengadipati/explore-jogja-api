package tourist

import (
	"testing"

	"pleco-api/internal/modules/event"
)

func TestCalculateEventQualityFullScore(t *testing.T) {
	e := &AIGenerateEventResponse{
		Title:            "Festival Seni Jogja",
		TitleEn:          "Jogja Arts Festival",
		Organizer:        "Panitia Festival",
		TicketPrice:      "Rp50.000",
		StartDate:        "2026-09-01",
		EndDate:          "2026-09-05",
		MaxAttendees:     "5000",
		Latitude:         "-7.7970",
		Longitude:        "110.3695",
		Description:      "A vibrant 5-day celebration of traditional Javanese arts, modern installations, and international performers gathered in the heart of Yogyakarta. Visitors can enjoy traditional dance performances, craft workshops, and local food stalls. The festival runs daily from 10am to 10pm with free shuttle service from the main square.",
		DescriptionEn:    "A vibrant 5-day celebration of traditional Javanese arts, modern installations, and international performers gathered in the heart of Yogyakarta. Visitors can enjoy traditional dance performances, craft workshops, and local food stalls. The festival runs daily from 10am to 10pm with free shuttle service from the main square.",
		SeoTitle:         "Festival Seni Jogja 2026",
		SeoTitleEn:       "Jogja Arts Festival 2026",
		SeoDescription:   "Ikutan Festival Seni Jogja 2026 di pusat kota Yogyakarta.",
		SeoDescriptionEn: "Join the Jogja Arts Festival 2026 in central Yogyakarta.",
		SeoKeywords:      "festival seni jogja, event jogja, yogyakarta",
		SeoKeywordsEn:    "jogja arts festival, jogja event, yogyakarta",
	}
	qs := calculateEventQuality(e)
	if qs.Total != qs.Max {
		t.Fatalf("expected perfect score %d, got %d (verdict %s)", qs.Max, qs.Total, qs.Verdict)
	}
	if qs.Verdict != event.VerdictExcellent {
		t.Fatalf("expected EXCELLENT, got %s", qs.Verdict)
	}
}

func TestCalculateEventQualityPoor(t *testing.T) {
	e := &AIGenerateEventResponse{}
	qs := calculateEventQuality(e)
	if qs.Total != 0 {
		t.Fatalf("expected 0 for empty, got %d", qs.Total)
	}
	if qs.Verdict != event.VerdictNeedsWork {
		t.Fatalf("expected NEEDS WORK, got %s", qs.Verdict)
	}
}

func TestEventFactDensityScore(t *testing.T) {
	e := &AIGenerateEventResponse{
		Title:          "Event",
		TitleEn:        "Event",
		Organizer:      "Org",
		TicketPrice:    "Rp10.000",
		StartDate:      "2026-01-01",
		EndDate:        "2026-01-02",
		MaxAttendees:   "100",
		Latitude:       "-7.7",
		Longitude:      "110.3",
		Description:    "Lorem ipsum dolor sit amet, consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi aliquip ex ea commodo consequat.",
		DescriptionEn:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi aliquip ex ea commodo consequat.",
		SeoTitle:       "SEO Title",
		SeoDescription: "SEO desc",
		SeoKeywords:    "kw",
	}
	if got := eventFactDensityScore(e); got != 13 {
		t.Fatalf("expected 13, got %d", got)
	}
}
