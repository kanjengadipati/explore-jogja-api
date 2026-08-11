package event

import (
	"math"
	"strconv"
	"strings"
)

// ─── Score types ──────────────────────────────────────────────────────────────
// Mirrors internal/modules/destination/quality.go — same verdict thresholds and
// 100-point total, adapted to the fields that actually exist on Event (no story,
// facilities, travel_tips, weather, or faqs on this model).

const (
	VerdictExcellent = "EXCELLENT"  // >= 80
	VerdictGood      = "GOOD"       // >= 60
	VerdictNeedsWork = "NEEDS WORK" // < 60

	// PublishScoreGate mirrors destination.PublishScoreGate. Not wired as a hard
	// publish block yet (events don't have a draft/review content lifecycle like
	// destinations do) — kept as a named constant so a future gate can reuse it
	// instead of a magic number.
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

func hasText(s string) bool             { return strings.TrimSpace(s) != "" }
func longEnough(s string, n int) bool   { return len(strings.TrimSpace(s)) >= n }
func validCoord(f float64) bool         { return !math.IsNaN(f) && !math.IsInf(f, 0) && f != 0 }
func countArr(raw JSONArr) int          { return len(raw) }

func imgCount(main string, gallery JSONArr) int {
	count := 0
	if hasText(main) {
		count++
	}
	for _, v := range gallery {
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

// ─── Rubric ───────────────────────────────────────────────────────────────────

// CalculateScore computes the deterministic 100-point content quality rubric
// for an event. Category weights mirror destination.CalculateScore's split
// (Identity 10 / Content 25 / Practical 15 / Media 20 / SEO 15 / Rich 15),
// with items substituted for fields that exist on Event.
func CalculateScore(e *Event) QualityScore {
	var cats []ScoreCategory

	// ── 1. Identity (10) ─────────────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "Title", Max: 2, Points: boolPts(hasText(e.Title), 2)},
			{Label: "Category", Max: 2, Points: boolPts(hasText(e.Category), 2)},
			{Label: "Location", Max: 2, Points: boolPts(hasText(e.Location), 2)},
			{Label: "Organizer", Max: 2, Points: boolPts(hasText(e.Organizer), 2)},
			{Label: "Coordinates", Max: 2, Points: boolPts(validCoord(e.Latitude) && validCoord(e.Longitude), 2)},
		}
		cats = append(cats, finishCat("identity", "Identity", 10, items))
	}

	// ── 2. Content (25) ──────────────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "Description (≥200 chars)", Max: 13, Points: boolPts(longEnough(e.Description, 200), 13)},
			{Label: "Description EN (≥200 chars)", Max: 12, Points: boolPts(longEnough(e.DescriptionEn, 200), 12)},
		}
		cats = append(cats, finishCat("content", "Content", 25, items))
	}

	// ── 3. Practical Info (15) ───────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "Ticket price", Max: 5, Points: boolPts(hasText(e.TicketPrice), 5)},
			{Label: "Start date", Max: 5, Points: boolPts(hasText(e.StartDate), 5)},
			{Label: "End date", Max: 5, Points: boolPts(hasText(e.EndDate), 5)},
		}
		cats = append(cats, finishCat("practical", "Practical Info", 15, items))
	}

	// ── 4. Media (20) ────────────────────────────────────────────────────────
	{
		n := imgCount(e.ImageURL, e.Images)
		imgPts := 0
		if n >= 3 {
			imgPts = 10
		} else if n >= 1 {
			imgPts = 5
		}
		items := []ScoreItem{
			{Label: "Gallery (≥3 images)", Max: 10, Points: imgPts, Detail: strconv.Itoa(n) + " images"},
			{Label: "OG image", Max: 5, Points: boolPts(hasText(e.OgImageUrl), 5)},
			{Label: "Video", Max: 5, Points: boolPts(hasText(e.VideoURL), 5)},
		}
		cats = append(cats, finishCat("media", "Media", 20, items))
	}

	// ── 5. SEO (15) ──────────────────────────────────────────────────────────
	{
		items := []ScoreItem{
			{Label: "SEO title", Max: 3, Points: boolPts(hasText(e.SeoTitle), 3)},
			{Label: "SEO description", Max: 3, Points: boolPts(hasText(e.SeoDescription), 3)},
			{Label: "SEO keywords", Max: 3, Points: boolPts(hasText(e.SeoKeywords), 3)},
			{Label: "SEO title EN", Max: 2, Points: boolPts(hasText(e.SeoTitleEn), 2)},
			{Label: "SEO description EN", Max: 2, Points: boolPts(hasText(e.SeoDescriptionEn), 2)},
			{Label: "SEO keywords EN", Max: 2, Points: boolPts(hasText(e.SeoKeywordsEn), 2)},
		}
		cats = append(cats, finishCat("seo", "SEO", 15, items))
	}

	// ── 6. Rich Content (15) ─────────────────────────────────────────────────
	{
		nHighlights := countArr(e.Highlights)
		items := []ScoreItem{
			{Label: "Highlights (≥3)", Max: 10, Points: boolPts(nHighlights >= 3, 10), Detail: strconv.Itoa(nHighlights) + " items"},
			{Label: "Max attendees set", Max: 5, Points: boolPts(e.MaxAttendees > 0, 5)},
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

func ApplyScoreToEvent(e *Event) {
	qs := CalculateScore(e)
	e.ContentScore = qs.Total
	e.ContentVerdict = qs.Verdict
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
