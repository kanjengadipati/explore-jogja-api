package adcampaign

import (
	"time"
)

// ─── Self-service pricing & approval policy ──────────────────────────────────
//
// These values back the business-portal self-service ad flow. They live here so
// the backend is the source of truth for price and the initial payment status;
// the admin UI mirrors the price list in src/lib/adPlacements.ts for display.
//
// PRICES ARE MONTHLY (IDR). Tweak them here; do NOT fork pricing logic in the
// frontend.

// MonthlyPrices is the flat monthly rate per sellable placement (IDR).
var MonthlyPrices = map[string]float64{
	PlacementHomepageHeroAICard:     300000,
	PlacementHomepageHeroTrending:   250000,
	PlacementHomepageCategoryBanner: 350000,
	PlacementListingTop:             250000,
	PlacementListingNative:          200000,
	PlacementDestinationDetail:      400000,

	PlacementEcosystemStay:       300000,
	PlacementEcosystemEat:        250000,
	PlacementEcosystemExperience: 250000,
	PlacementEcosystemShop:       200000,
	PlacementEcosystemMove:       200000,
	PlacementEcosystemGuide:      200000,
}

func IsSellablePlacement(placement string) bool {
	_, ok := MonthlyPrices[placement]
	return ok
}

// InitialPaymentStatus returns the starting status for a newly created ad
// campaign. Every sellable placement starts under review; staff must approve
// the creative before payment is opened.
func InitialPaymentStatus(placement string) string {
	if IsSellablePlacement(placement) {
		return PaymentStatusPendingReview
	}
	return PaymentStatusPendingPayment
}

// PriceFor computes the flat fee for a campaign period: the monthly rate times
// the number of calendar months covered (partial months round up, min 1 month).
// Zero dates are treated as a single month.
func PriceFor(placement string, start, end time.Time) float64 {
	monthly, ok := MonthlyPrices[placement]
	if !ok {
		return 0
	}
	startZero := start.IsZero()
	endZero := end.IsZero()
	if startZero || endZero {
		return monthly
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		return monthly
	}
	months := (days + 29) / 30
	if months < 1 {
		months = 1
	}
	return monthly * float64(months)
}

// PlacementLabel is the human-readable Indonesian label used in emails.
var PlacementLabel = map[string]string{
	PlacementHomepageHeroAICard:     "Hero AI Pick",
	PlacementHomepageHeroTrending:   "Hero Trending Now",
	PlacementHomepageCategoryBanner: "Banner Kategori Homepage",
	PlacementListingTop:             "Grid Destinasi Populer",
	PlacementListingNative:          "Native Ad Festival & Destinasi",
	PlacementDestinationDetail:      "Sponsorship Halaman Destinasi",

	PlacementEcosystemStay:       "Rel Rekomendasi — Menginap",
	PlacementEcosystemEat:        "Rel Rekomendasi — Kuliner",
	PlacementEcosystemExperience: "Rel Rekomendasi — Vibe & Aktivitas",
	PlacementEcosystemShop:       "Rel Rekomendasi — Belanja",
	PlacementEcosystemMove:       "Rel Rekomendasi — Transport",
	PlacementEcosystemGuide:      "Rel Rekomendasi — Guide Lokal",
}

func PlacementName(placement string) string {
	if name, ok := PlacementLabel[placement]; ok {
		return name
	}
	return placement
}
