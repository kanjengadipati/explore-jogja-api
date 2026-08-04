package scraper

// ScrapedEvent represents an event parsed from an external source.
type ScrapedEvent struct {
	ExternalID    string
	Title         string
	Description   string
	Location      string
	StartDate     string
	EndDate       string
	ImageURL      string
	Category      string
	Latitude      float64
	Longitude     float64
	TicketPrice   string
	Organizer     string
	Highlights    []string
	DestinationID string
	VideoURL      string
	Source        string
}

// ScrapedDestination represents a destination parsed from an external source.
type ScrapedDestination struct {
	ExternalID  string
	Name        string
	Tagline     string
	Category    string
	Location    string
	SubRegion   string
	Images      []string
	Description string
	Story       string
	TicketPrice string
	Latitude    float64
	Longitude   float64
	VideoURL    string
	Source      string
}

// ScrapeResult holds the outcome of a single scraper run. New items never go
// straight into the live tables — they are queued for human review instead —
// so there are no "inserted" counters, only updated/staged.
type ScrapeResult struct {
	Source              string
	EventsUpdated       int
	EventsStaged        int
	DestinationsUpdated int
	DestinationsStaged  int
	Errors              []string
}

// Scraper defines the interface each source must implement.
type Scraper interface {
	Name() string
	ScrapeEvents() ([]ScrapedEvent, error)
	ScrapeDestinations() ([]ScrapedDestination, error)
}
