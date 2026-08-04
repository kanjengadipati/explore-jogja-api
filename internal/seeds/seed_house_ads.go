package seeds

import (
	"fmt"
	"log"

	"pleco-api/internal/modules/adcampaign"

	"gorm.io/gorm"
)

// SeedHouseAds seeds the self-promo House Ads (the "Pasang Iklan" slots).
// It is the source of truth for the default house_ads rows (mirrors
// migrations/seed_house_ads_self_promo.sql). It only creates missing
// placements or repairs rows with an empty headline, so real
// admin/editorial content is never overwritten.
func SeedHouseAds(db *gorm.DB) {
	mustHaveDB(db)

	ads := []adcampaign.HouseAd{
		{
			ExternalID: "house_ad_promo_hero",
			Placement:  adcampaign.PlacementHomepageHeroAICard,
			Headline:   "Ribuan wisatawan lihat halaman ini tiap hari",
			HeadlineEn: "Thousands of travelers see this page every day",
			Subline:    "Pasang bisnismu di posisi paling depan Jogjagem.",
			SublineEn:  "Put your business at the very front of Jogjagem.",
			CTALabel:   "Pasang Iklan",
			CTALabelEn: "Advertise Now",
			ImageURL:   "/images/house-ads/promo-hero.jpg",
			TargetURL:  "/ads?placement=homepage_hero_aicard",
			IsEnabled:  true,
		},
		{
			ExternalID: "house_ad_promo_listing_top",
			Placement:  adcampaign.PlacementListingTop,
			Headline:   "Bisnismu bisa tampil paling atas di sini",
			HeadlineEn: "Your business can appear at the very top here",
			Subline:    "Slot ini kosong sekarang — jadi yang pertama dilihat pengunjung.",
			SublineEn:  "This slot is empty right now — be the first thing visitors see.",
			CTALabel:   "Pasang Iklan",
			CTALabelEn: "Advertise Now",
			ImageURL:   "/images/house-ads/promo-listing-top.jpg",
			TargetURL:  "/ads?placement=listing_top",
			IsEnabled:  true,
		},
		{
			ExternalID: "house_ad_promo_listing_native",
			Placement:  adcampaign.PlacementListingNative,
			Headline:   "Muncul natural di antara listing populer",
			HeadlineEn: "Show up naturally among popular listings",
			Subline:    "Terintegrasi langsung di alur pencarian pengunjung.",
			SublineEn:  "Integrated right into the visitor's search flow.",
			CTALabel:   "Pasang Iklan",
			CTALabelEn: "Advertise Now",
			ImageURL:   "/images/house-ads/promo-listing-native.jpg",
			TargetURL:  "/ads?placement=listing_native",
			IsEnabled:  true,
		},
		{
			ExternalID: "house_ad_promo_destination",
			Placement:  adcampaign.PlacementDestinationDetail,
			Headline:   "Promosikan usahamu ke pengunjung destinasi ini",
			HeadlineEn: "Promote your business to visitors of this destination",
			Subline:    "Tampil tepat saat wisatawan sedang merencanakan kunjungan.",
			SublineEn:  "Shows exactly when travelers are planning their visit.",
			CTALabel:   "Pasang Iklan",
			CTALabelEn: "Advertise Now",
			ImageURL:   "/images/house-ads/promo-destination.jpg",
			TargetURL:  "/ads?placement=destination_detail",
			IsEnabled:  true,
		},
	}

	inserted, updated := 0, 0
	for _, ad := range ads {
		var existing adcampaign.HouseAd
		err := db.Where("placement = ?", ad.Placement).First(&existing).Error
		switch {
		case err == gorm.ErrRecordNotFound:
			if dbErr := db.Create(&ad).Error; dbErr != nil {
				log.Printf("Failed to seed house ad %s: %v", ad.Placement, dbErr)
			} else {
				inserted++
			}
		case err != nil:
			log.Printf("Failed to check house ad %s: %v", ad.Placement, err)
		case existing.Headline == "":
			ad.ID = existing.ID
			ad.CreatedAt = existing.CreatedAt
			if dbErr := db.Save(&ad).Error; dbErr != nil {
				log.Printf("Failed to repair house ad %s: %v", ad.Placement, dbErr)
			} else {
				updated++
			}
		default:
			// Placement already has real content — leave it alone.
		}
	}
	fmt.Printf("House ads seeding done: %d inserted, %d updated\n", inserted, updated)
}
