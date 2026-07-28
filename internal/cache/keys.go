package cache

import "time"

// Well-known Redis key constants used across modules.

// KeyAITrendingResponse returns a locale-scoped cache key for the full AI
// trending response payload ([]AITrendingItem).
func KeyAITrendingResponse(locale string) string {
	return "ai:trending:response:" + locale
}

// KeyAITrendingIDs returns a locale-scoped cache key for the destination
// external-IDs elected as trending. Written alongside KeyAITrendingResponse
// and read by the destination/event badge logic.
func KeyAITrendingIDs(locale string) string {
	return "ai:trending:destination_ids:" + locale
}

const (
	// TTLAITrending is how long both trending cache keys stay valid.
	// 7 days = cache resets once a week automatically.
	TTLAITrending = 7 * 24 * time.Hour

	// KeyDestinationsAllPrefix is the prefix for caching all destinations by locale.
	KeyDestinationsAllPrefix = "destinations:all:"
	// KeyDestinationsCategoryPrefix is the prefix for caching destinations by category and locale.
	KeyDestinationsCategoryPrefix = "destinations:category:"
	// KeyDestinationsIDPrefix is the prefix for caching a destination by ID and locale.
	KeyDestinationsIDPrefix = "destinations:id:"
	// TTLDestinations is how long destination cache stays valid. 7 days.
	TTLDestinations = 7 * 24 * time.Hour

	// KeyEventsAll is the cache key for all events.
	KeyEventsAll = "events:all"
	// KeyEventsIDPrefix is the prefix for caching an event by ID.
	KeyEventsIDPrefix = "events:id:"
	// TTLEvents is how long event cache stays valid. 7 days.
	TTLEvents = 7 * 24 * time.Hour
)


