package seeds_test

import (
	"os"
	"testing"

	"pleco-api/internal/modules/adcampaign"
	"pleco-api/internal/seeds"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func openHouseAdsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func setupHouseAdsTempTable(t *testing.T, db *gorm.DB) *gorm.DB {
	t.Helper()

	tx := db.Begin()
	require.NoError(t, tx.Error)

	require.NoError(t, tx.Exec(`
		CREATE TEMP TABLE house_ads (
			id SERIAL PRIMARY KEY,
			external_id TEXT NOT NULL UNIQUE,
			placement TEXT NOT NULL UNIQUE,
			headline TEXT NOT NULL,
			headline_en TEXT,
			subline TEXT,
			subline_en TEXT,
			cta_label TEXT NOT NULL,
			cta_label_en TEXT,
			image_url TEXT,
			target_url TEXT NOT NULL,
			is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		)
	`).Error)

	t.Cleanup(func() { _ = tx.Rollback().Error })

	return tx
}

func TestSeedHouseAds_EmptyTable(t *testing.T) {
	db := openHouseAdsTestDB(t)
	tx := setupHouseAdsTempTable(t, db)

	seeds.SeedHouseAds(tx)

	var count int64
	require.NoError(t, tx.Model(&adcampaign.HouseAd{}).Count(&count).Error)
	require.Equal(t, int64(4), count)

	for _, placement := range []string{
		adcampaign.PlacementHomepageHeroAICard,
		adcampaign.PlacementListingTop,
		adcampaign.PlacementListingNative,
		adcampaign.PlacementDestinationDetail,
	} {
		var ad adcampaign.HouseAd
		require.NoError(t, tx.Where("placement = ?", placement).First(&ad).Error, "placement %s should exist", placement)
		require.NotEmpty(t, ad.Headline, "headline for %s", placement)
		require.NotEmpty(t, ad.HeadlineEn, "headline_en for %s", placement)
		require.NotEmpty(t, ad.CTALabelEn, "cta_label_en for %s", placement)
		require.True(t, ad.IsEnabled, "is_enabled for %s", placement)
	}
}

func TestSeedHouseAds_SkipsExistingContent(t *testing.T) {
	db := openHouseAdsTestDB(t)
	tx := setupHouseAdsTempTable(t, db)

	existing := adcampaign.HouseAd{
		ExternalID: "custom-id",
		Placement:  adcampaign.PlacementHomepageHeroAICard,
		Headline:   "Konten asli admin",
		Subline:    "Konten subline asli",
		CTALabel:   "Coba Sekarang",
		ImageURL:   "/images/custom.jpg",
		TargetURL:  "/custom",
		IsEnabled:  true,
	}
	require.NoError(t, tx.Create(&existing).Error)

	seeds.SeedHouseAds(tx)

	var ad adcampaign.HouseAd
	require.NoError(t, tx.Where("placement = ?", adcampaign.PlacementHomepageHeroAICard).First(&ad).Error)
	require.Equal(t, "custom-id", ad.ExternalID, "external_id must not be overwritten")
	require.Equal(t, "Konten asli admin", ad.Headline, "headline must not be overwritten")
	require.Empty(t, ad.HeadlineEn, "bilingual fields left alone on existing content")
}

func TestSeedHouseAds_RepairsEmptyRow(t *testing.T) {
	db := openHouseAdsTestDB(t)
	tx := setupHouseAdsTempTable(t, db)

	stale := adcampaign.HouseAd{
		ExternalID: "placeholder",
		Placement:  adcampaign.PlacementListingTop,
		Headline:   "",
		CTALabel:   "",
		TargetURL:  "/placeholder",
	}
	require.NoError(t, tx.Create(&stale).Error)

	seeds.SeedHouseAds(tx)

	var ad adcampaign.HouseAd
	require.NoError(t, tx.Where("placement = ?", adcampaign.PlacementListingTop).First(&ad).Error)
	require.Equal(t, adcampaign.PlacementListingTop, ad.Placement)
	require.NotEmpty(t, ad.Headline, "empty headline should be repaired")
	require.NotEmpty(t, ad.HeadlineEn, "headline_en should be repaired")
	require.Equal(t, "/ads?placement=listing_top", ad.TargetURL)
}
