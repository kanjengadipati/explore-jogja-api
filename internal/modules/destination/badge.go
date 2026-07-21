package destination

import (
	"strings"
)

// BadgeType represents the possible badge labels for a destination.
type BadgeType string

const (
	BadgeTrending       BadgeType = "trending"
	BadgeHiddenGem      BadgeType = "hidden_gem"
	BadgeBestForHealing BadgeType = "best_for_healing"
	BadgeInstagramable  BadgeType = "instagramable"
	BadgeSunsetSpot     BadgeType = "sunset_spot"
	BadgeCultural       BadgeType = "cultural"
	BadgePerfectMorning BadgeType = "perfect_morning"
	BadgeAdventure      BadgeType = "adventure"
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

	// Trending: AI-selected as trending today
	if trendingIDs[d.ExternalID] {
		badges = append(badges, BadgeTrending)
	}

	// Hidden Gem: Rating >= 4.5 + review_count < 2500
	if d.Rating >= 4.5 && d.ReviewCount < 2500 {
		badges = append(badges, BadgeHiddenGem)
	}

	// Best for Healing: Category: nature + rating >= 4.3
	if cat == "nature" && d.Rating >= 4.3 {
		badges = append(badges, BadgeBestForHealing)
	}

	// Instagramable: Category: hidden-gem atau beach + rating >= 4.4
	if (cat == "hidden-gem" || cat == "beach") && d.Rating >= 4.4 {
		badges = append(badges, BadgeInstagramable)
	}

	// Sunset Spot: Category: beach + bestTime mengandung "sore" atau "sunset"
	if cat == "beach" && (strings.Contains(bestTime, "sore") || strings.Contains(bestTime, "sunset")) {
		badges = append(badges, BadgeSunsetSpot)
	}

	// Cultural: Category: heritage
	if cat == "heritage" {
		badges = append(badges, BadgeCultural)
	}

	// Perfect Morning: Category: nature atau heritage + bestTime mengandung "pagi"
	if (cat == "nature" || cat == "heritage") && strings.Contains(bestTime, "pagi") {
		badges = append(badges, BadgePerfectMorning)
	}

	// Adventure: Category: adventure + rating >= 4.0
	if cat == "adventure" && d.Rating >= 4.0 {
		badges = append(badges, BadgeAdventure)
	}

	return badges
}

// badgePriority defines display priority order (highest first).
var badgePriority = []BadgeType{
	BadgeTrending,
	BadgeHiddenGem,
	BadgeBestForHealing,
	BadgeInstagramable,
	BadgeSunsetSpot,
	BadgeCultural,
	BadgePerfectMorning,
	BadgeAdventure,
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
