package destination

import (
	"encoding/json"
	"math"
	"strings"
)

// ─── Score types ──────────────────────────────────────────────────────────────

const (
	VerdictExcellent  = "EXCELLENT"  // >= 80
	VerdictGood       = "GOOD"       // >= 60
	VerdictNeedsWork  = "NEEDS WORK" // < 60

	// PublishScoreGate is the minimum score required to set status=published via API.
	PublishScoreGate = 60
)

type ScoreItem struct {
	Label  string `json:"label"`
	Points int    `json:"points"`
	Max    int    `json:"max"`
	Detail string `json:"detail,omitempty"`
}

type ScoreCategory struct {
	Key   string      `json:"key"`
	Label string      `json:"label"`
	Score int         `json:"score"`
	Max   int         `json:"max"`
	Items []ScoreItem `json:"items"`
}

type QualityScore struct {
	Total      int             `json:"total"`
	Max        int             `json:"max"`
	Verdict    string          `json:"verdict"`
	Categories []ScoreCategory `json:"categories"`
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func hasText(s string) bool  { return strings.TrimSpace(s) != "" }
func longEnough(s string, n int) bool { return len(strings.TrimSpace(s)) >= n }

func parseJSONArr(raw JSONArr) []interface{} {
	if raw == nil {
		return nil
	}
	return []interface{}(raw)
}

func countArr(raw JSONArr) int { return len(parseJSONArr(raw)) }

func validCoord(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) && f != 0 }

func imgCount(images JSONArr) int {
	// images is []interface{} where each entry is either a string URL or a map {"url":...}
	count := 0
	for _, v := range images {
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				count++
			}
		case map[string]interface{}:
			if u, ok := t["url"].(string); ok && strings.TrimSpace(u) != "" {
				count++
			}
		}
	}
	return count
}

func weatherFilled(w JSONMap) bool {
	if w == nil {
		return false
	}
	temp, _ := w["temp"].(string)
	cond, _ := w["condition"].(string)
	return hasText(temp) && hasText(cond)
}

// ─── Rubric ───────────────────────────────────────────────────────────────────

// CalculateScore computes the deterministic 100-point content quality rubric
// for a destination. Mirrors content-score.ts in jogjagem-admin exactly.
func CalculateScore(d *Destination) QualityScore {
	var cats []ScoreCategory

	// ── 1. Identity (10) ─────────────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "Name",        Max: 2, Points: boolPts(hasText(d.Name), 2)},
			{Label: "Tagline",     Max: 2, Points: boolPts(hasText(d.Tagline), 2)},
			{Label: "Category",    Max: 2, Points: boolPts(hasText(d.Category), 2)},
			{Label: "Location",    Max: 2, Points: boolPts(hasText(d.Location), 2)},
			{Label: "Sub-region",  Max: 1, Points: boolPts(hasText(d.SubRegion), 1)},
			{Label: "Coordinates", Max: 1, Points: boolPts(validCoord(d.Latitude) && validCoord(d.Longitude), 1)},
		}
		cats = append(cats, finishCat("identity", "Identity", 10, items))
	}

	// ── 2. Content (25) ──────────────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "Description (≥200 chars)",    Max: 6, Points: boolPts(longEnough(d.Description, 200), 6)},
			{Label: "Description EN (≥200 chars)", Max: 6, Points: boolPts(longEnough(d.DescriptionEn, 200), 6)},
			{Label: "Story (≥300 chars)",          Max: 7, Points: boolPts(longEnough(d.Story, 300), 7)},
			{Label: "Story EN (≥300 chars)",       Max: 6, Points: boolPts(longEnough(d.StoryEn, 300), 6)},
		}
		cats = append(cats, finishCat("content", "Content", 25, items))
	}

	// ── 3. Practical Info (15) ───────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "Ticket price",   Max: 4, Points: boolPts(hasText(d.TicketPrice), 4)},
			{Label: "Opening hours",  Max: 4, Points: boolPts(hasText(d.OpeningHours), 4)},
			{Label: "Best time (ID)", Max: 4, Points: boolPts(hasText(d.BestTime), 4)},
			{Label: "Best time (EN)", Max: 3, Points: boolPts(hasText(d.BestTimeEn), 3)},
		}
		cats = append(cats, finishCat("practical", "Practical Info", 15, items))
	}

	// ── 4. Media (20) ────────────────────────────────────────────────────────
	{
		n := imgCount(d.Images)
		imgPts := 0
		if n >= 3 {
			imgPts = 10
		} else if n >= 1 {
			imgPts = 5
		}
		imgDetail := itoa(n) + " image"
		if n != 1 {
			imgDetail += "s"
		}
		items := []ScoreItem{
			{Label: "Gallery (≥3 images)", Max: 10, Points: imgPts, Detail: imgDetail},
			{Label: "OG image",            Max: 5,  Points: boolPts(hasText(d.OgImageUrl), 5)},
			{Label: "Video",               Max: 5,  Points: boolPts(hasText(d.VideoURL), 5)},
		}
		cats = append(cats, finishCat("media", "Media", 20, items))
	}

	// ── 5. SEO (15) ──────────────────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "SEO title",          Max: 3, Points: boolPts(hasText(d.SeoTitle), 3)},
			{Label: "SEO description",    Max: 3, Points: boolPts(hasText(d.SeoDescription), 3)},
			{Label: "SEO keywords",       Max: 3, Points: boolPts(hasText(d.SeoKeywords), 3)},
			{Label: "SEO title EN",       Max: 2, Points: boolPts(hasText(d.SeoTitleEn), 2)},
			{Label: "SEO description EN", Max: 2, Points: boolPts(hasText(d.SeoDescriptionEn), 2)},
			{Label: "SEO keywords EN",    Max: 2, Points: boolPts(hasText(d.SeoKeywordsEn), 2)},
		}
		cats = append(cats, finishCat("seo", "SEO", 15, items))
	}

	// ── 6. Rich Content (15) ─────────────────────────────────────────────────
	{
		nFAQ  := countArr(d.FAQs)
		nTips := countArr(d.TravelTips)
		nFacs := countArr(d.Facilities)
		items := []ScoreItem{
			{Label: "FAQs (≥3)",        Max: 5, Points: boolPts(nFAQ >= 3, 5),  Detail: itoa(nFAQ) + " items"},
			{Label: "Travel tips (≥3)", Max: 5, Points: boolPts(nTips >= 3, 5), Detail: itoa(nTips) + " items"},
			{Label: "Facilities (≥3)",  Max: 3, Points: boolPts(nFacs >= 3, 3), Detail: itoa(nFacs) + " items"},
			{Label: "Weather",          Max: 2, Points: boolPts(weatherFilled(d.Weather), 2)},
		}
		cats = append(cats, finishCat("rich", "Rich Content", 15, items))
	}

	total := 0
	maxTotal := 0
	for _, c := range cats {
		total += c.Score
		maxTotal += c.Max
	}

	verdict := VerdictNeedsWork
	if total >= 80 {
		verdict = VerdictExcellent
	} else if total >= 60 {
		verdict = VerdictGood
	}

	return QualityScore{
		Total:      total,
		Max:        maxTotal,
		Verdict:    verdict,
		Categories: cats,
	}
}

// PersistScore is kept for future use (direct DB update by external callers).
func PersistScore(_ *Destination) {}

// AutoFillGoogleMapsURL sets GoogleMapsURL from coordinates if it's empty and
// valid coordinates are available. Safe to call unconditionally — no-ops when
// GoogleMapsURL is already set or coordinates are zero.
func AutoFillGoogleMapsURL(d *Destination) {
	if d.GoogleMapsURL != "" {
		return // already set — don't overwrite
	}
	if d.Latitude == 0 || d.Longitude == 0 {
		return // no coordinates — can't build URL
	}
	d.GoogleMapsURL = fmt.Sprintf(
		"https://www.google.com/maps?q=%.6f,%.6f",
		d.Latitude, d.Longitude,
	)
}
func ApplyScoreToDestination(d *Destination) {
	qs := CalculateScore(d)
	d.ContentScore = qs.Total
	d.ContentVerdict = qs.Verdict
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func boolPts(cond bool, pts int) int {
	if cond {
		return pts
	}
	return 0
}

func finishCat(key, label string, max int, items []ScoreItem) ScoreCategory {
	score := 0
	for _, it := range items {
		score += it.Points
	}
	return ScoreCategory{Key: key, Label: label, Score: score, Max: max, Items: items}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
