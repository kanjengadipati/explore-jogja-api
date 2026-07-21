package cache

import "time"

// Well-known Redis key constants used across modules.

const (
	// KeyAITrendingResponse caches the full AI trending response payload
	// ([]AITrendingItem) returned to the client. The tourist/handler checks
	// this before calling the AI provider, saving token costs.
	KeyAITrendingResponse = "ai:trending:response"

	// KeyAITrendingIDs caches only the destination external-IDs that the AI
	// elected as trending. Written alongside KeyAITrendingResponse and read by
	// the destination/handler badge logic so every destination list/detail call
	// can mark the trending badge without an extra AI round-trip.
	KeyAITrendingIDs = "ai:trending:destination_ids"

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


