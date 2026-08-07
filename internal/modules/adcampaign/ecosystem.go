package adcampaign

import (
	"math"
	"strconv"
)

// EcosystemCard is the public response shape for a sponsored entry in the
// destination detail "Rekomendasi Kebutuhan Traveler" rails. It deliberately
// mirrors partner.PublicPartnerResponse so the web app's mapBePartner keeps
// working unchanged, plus business_id (claim CTA) and the campaign's target_url.
//
// The card id is the campaign external id: impression/click tracking therefore
// goes through the ads campaign counters instead of the retired partner module.
type EcosystemCard struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Location    string  `json:"location"`
	Address     string  `json:"address"`
	Image       string  `json:"image"`
	Rating      float64 `json:"rating"`
	Price       string  `json:"price"`
	Distance    string  `json:"distance"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	IsSponsored bool    `json:"is_sponsored"`
	SponsorTier int     `json:"sponsor_tier"`
	Status      string  `json:"status"`
	BusinessID  string  `json:"business_id"`
	TargetURL   string  `json:"target_url"`
}

// GetEcosystem serves the sponsored cards for the ecosystem rails of a
// destination. Campaigns are resolved through the ads system (placement
// ecosystem_*, payment_status paid, active within start/end, targeting the
// destination), then enriched with the promoted listing's card data. The
// distance field is computed from the destination coordinates when available.
func (s *Service) GetEcosystem(destinationID string) ([]EcosystemCard, error) {
	campaigns, err := s.Repo.FindActiveEcosystemCandidates(destinationID)
	if err != nil {
		return nil, err
	}

	destLat, destLng, hasDest := 0.0, 0.0, false
	if destinationID != "" {
		destLat, destLng, hasDest = s.Repo.FindDestinationCoords(destinationID)
	}

	cards := make([]EcosystemCard, 0, len(campaigns))
	for _, c := range campaigns {
		listing, err := s.Repo.FindEcosystemListing(c.ListingType, c.ListingExternalID)
		if err != nil || listing == nil || listing.Name == "" {
			// Listing missing/unlinked — skip instead of rendering an empty card.
			continue
		}

		distance := ""
		if hasDest && listing.Latitude != 0 && listing.Longitude != 0 {
			distance = formatDistance(haversineKm(destLat, destLng, listing.Latitude, listing.Longitude))
		}

		businessID := ""
		if c.BusinessExternalID != nil {
			businessID = *c.BusinessExternalID
		}

		cards = append(cards, EcosystemCard{
			ID:          c.ExternalID,
			Name:        listing.Name,
			Description: listing.Description,
			Category:    c.ListingType,
			Address:     listing.Address,
			Image:       listing.Image,
			Rating:      listing.Rating,
			Price:       listing.Price,
			Distance:    distance,
			Phone:       listing.Phone,
			Website:     listing.Website,
			Latitude:    listing.Latitude,
			Longitude:   listing.Longitude,
			IsSponsored: true,
			SponsorTier: c.SortOrder + 1,
			Status:      "approved",
			BusinessID:  businessID,
			TargetURL:   c.TargetURL,
		})
	}
	return cards, nil
}

// haversineKm returns the great-circle distance in kilometres between two
// WGS84 coordinate pairs.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0
	rad := math.Pi / 180
	dLat := (lat2 - lat1) * rad
	dLng := (lng2 - lng1) * rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return earthRadiusKm * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// formatDistance renders a kilometre distance the way the ecosystem cards
// expect ("450 m" below 1 km, "X.X km" above).
func formatDistance(km float64) string {
	if km < 1 {
		return strconv.FormatFloat(km*1000, 'f', 0, 64) + " m"
	}
	return strconv.FormatFloat(km, 'f', 1, 64) + " km"
}
