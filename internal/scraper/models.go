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
	Source      string
}

// ScrapeResult holds the outcome of a single scraper run.
type ScrapeResult struct {
	Source           string
	EventsInserted   int
	EventsUpdated    int
	DestinationsInserted int
	DestinationsUpdated int
	Errors           []string
}

// Scraper defines the interface each source must implement.
type Scraper interface {
	Name() string
	ScrapeEvents() ([]ScrapedEvent, error)
	ScrapeDestinations() ([]ScrapedDestination, error)
}
