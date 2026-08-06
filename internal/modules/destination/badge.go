package destination

import (
	"strings"
)

// BadgeType represents the possible badge labels for a destination.
type BadgeType string

const (
	BadgeTrending        BadgeType = "trending"
	BadgeHiddenGem       BadgeType = "hidden_gem"
	BadgeBestForHealing  BadgeType = "best_for_healing"
	BadgeInstagramable   BadgeType = "instagramable"
	BadgeSunsetSpot      BadgeType = "sunset_spot"
	BadgeSunriseSpot     BadgeType = "sunrise_spot"
	BadgeCultural        BadgeType = "cultural"
	BadgePerfectMorning  BadgeType = "perfect_morning"
	BadgeAdventure       BadgeType = "adventure"
	BadgeCampingSpot     BadgeType = "camping_spot"
	BadgeBudgetFriendly  BadgeType = "budget_friendly"
	BadgeWaterfall       BadgeType = "waterfall"
	BadgeNightSpot       BadgeType = "night_spot"
	BadgePhotographerPick BadgeType = "photographer_pick"
)

// ResolveBadges returns all badges that apply to the given destination.
//
// trendingIDs is a set of destination external-IDs that the AI Trending
// endpoint has elected as trending today.  Pass nil or empty map when the
// value is unavailable (badge will simply not be assigned).
func ResolveBadges(d Destination, trendingIDs map[string]bool) []BadgeType {
	var badges []BadgeType

	cat := strings.ToLower(strings.TrimSpace(d.Category))
	bestTime := strings.ToLower(d.BestTime)
	tips := strings.ToLower(travelTipsString(d.TravelTips))
	ticketPrice := strings.ToLower(strings.TrimSpace(d.TicketPrice))
	name := strings.ToLower(strings.TrimSpace(d.Name))

	// Trending: AI-selected as trending today
	if trendingIDs[d.ExternalID] {
		badges = append(badges, BadgeTrending)
	}

	// Sunrise Spot: bestTime mengandung "sunrise" / "fajar" / "dawn"
	if strings.Contains(bestTime, "sunrise") || strings.Contains(bestTime, "fajar") || strings.Contains(bestTime, "dawn") {
		badges = append(badges, BadgeSunriseSpot)
	}

	// Sunset Spot: bestTime mengandung "sore" atau "sunset", tapi hanya untuk nature/beach
	if (cat == "nature" || cat == "beach") && (strings.Contains(bestTime, "sore") || strings.Contains(bestTime, "sunset")) {
		badges = append(badges, BadgeSunsetSpot)
	}

	// Camping Spot: facilities mengandung "camping" ATAU bestTime mengandung "camping" (fallback)
	if hasFacility(d.Facilities, "camping") || strings.Contains(bestTime, "camping") {
		badges = append(badges, BadgeCampingSpot)
	}

	// Waterfall: name mengandung "air terjun" / "curug" / "waterfall"
	if strings.Contains(name, "air terjun") || strings.Contains(name, "curug") || strings.Contains(name, "waterfall") {
		badges = append(badges, BadgeWaterfall)
	}

	// Instagramable: Category hidden-gem atau beach + rating >= 4.4
	if (cat == "hidden-gem" || cat == "beach") && d.Rating >= 4.4 {
		badges = append(badges, BadgeInstagramable)
	}

	// Perfect Morning: Category nature atau heritage + bestTime mengandung "pagi"
	if (cat == "nature" || cat == "heritage") && strings.Contains(bestTime, "pagi") {
		badges = append(badges, BadgePerfectMorning)
	}

	// Night Spot: bestTime mengandung "malam" / "night" / "stargaz"
	if strings.Contains(bestTime, "malam") || strings.Contains(bestTime, "night") || strings.Contains(bestTime, "stargaz") {
		badges = append(badges, BadgeNightSpot)
	}

	// Photographer Pick: travelTips mengandung "foto" / "photo" / "fotografi"
	if strings.Contains(tips, "foto") || strings.Contains(tips, "photo") || strings.Contains(tips, "fotografi") {
		badges = append(badges, BadgePhotographerPick)
	}

	// Best for Healing: Category nature + rating >= 4.3
	if cat == "nature" && d.Rating >= 4.3 {
		badges = append(badges, BadgeBestForHealing)
	}

	// Cultural: Category heritage
	if cat == "heritage" {
		badges = append(badges, BadgeCultural)
	}

	// Adventure: Category adventure + rating >= 4.0
	if cat == "adventure" && d.Rating >= 4.0 {
		badges = append(badges, BadgeAdventure)
	}

	// Hidden Gem: Rating >= 4.5 + review_count < 2500 (after category-specific badges)
	if d.Rating >= 4.5 && d.ReviewCount < 2500 {
		badges = append(badges, BadgeHiddenGem)
	}

	// Budget Friendly: ticket_price gratis / free
	if strings.Contains(ticketPrice, "gratis") || strings.Contains(ticketPrice, "free") {
		badges = append(badges, BadgeBudgetFriendly)
	}

	return badges
}

func travelTipsString(tips JSONArr) string {
	var parts []string
	for _, t := range tips {
		if s, ok := t.(string); ok {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " ")
}

func hasFacility(facilities JSONArr, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, v := range facilities {
		if s, ok := v.(string); ok {
			if strings.Contains(strings.ToLower(strings.TrimSpace(s)), target) {
				return true
			}
		}
	}
	return false
}

// badgePriority defines display priority order (highest first).
// Category-specific badges (Cultural, Adventure, etc.) are not in this list
// because they are always the primary badge for their category and don't need
// to compete with content-based badges for the "primary" slot.
var badgePriority = []BadgeType{
	BadgeTrending,
	BadgeSunsetSpot,
	BadgeSunriseSpot,
	BadgeCampingSpot,
	BadgeWaterfall,
	BadgeInstagramable,
	BadgePerfectMorning,
	BadgeNightSpot,
	BadgePhotographerPick,
	BadgeBestForHealing,
	BadgeHiddenGem,
	BadgeBudgetFriendly,
}

// PrimaryBadge returns the single highest-priority badge for card overlay display.
// Returns empty string if no badge applies.
func PrimaryBadge(d Destination, trendingIDs map[string]bool) BadgeType {
	all := ResolveBadges(d, trendingIDs)
	if len(all) == 0 {
		return ""
	}
	set := make(map[BadgeType]bool, len(all))
	for _, b := range all {
		set[b] = true
	}
	for _, p := range badgePriority {
		if set[p] {
			return p
		}
	}
	return all[0]
}
