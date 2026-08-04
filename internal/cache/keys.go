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

	// TTLEvents is how long event cache stays valid. 7 days.
	TTLEvents = 7 * 24 * time.Hour
)

// KeyEventsAll returns a locale-scoped cache key for the full events list.
// Events are localized (title/description) and badged per locale, so the
// cache must never be shared across locales.
func KeyEventsAll(locale string) string {
	return "events:all:" + locale
}

// KeyEventsID returns a locale-scoped cache key for a single event by ID.
func KeyEventsID(locale, id string) string {
	return "events:id:" + locale + ":" + id
}

// KeyEventsAllPrefix is the prefix for every locale variant of the events
// list. Used to invalidate the whole list on writes.
const KeyEventsAllPrefix = "events:all:"

// KeyEventsIDAllPrefix is the prefix for every locale variant of the single
// event cache. Used to invalidate event-by-ID caches on writes.
const KeyEventsIDAllPrefix = "events:id:"
