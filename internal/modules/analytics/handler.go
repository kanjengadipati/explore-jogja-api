package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pleco-api/internal/httpx"
)

type Handler struct {
	DB *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}

type OverviewResponse struct {
	TotalDestinations   int64   `json:"total_destinations"`
	TotalEvents         int64   `json:"total_events"`
	TotalUsers          int64   `json:"total_users"`
	TotalReviews        int64   `json:"total_reviews"`
	TotalStories        int64   `json:"total_stories"`
	TotalPartners       int64   `json:"total_partners"`
	TotalHotels         int64   `json:"total_hotels"`
	TotalRestaurants    int64   `json:"total_restaurants"`
	TotalGuides         int64   `json:"total_guides"`
	TotalSouvenirs      int64   `json:"total_souvenirs"`
	TotalRentals        int64   `json:"total_rentals"`
	TotalArticles       int64   `json:"total_articles"`
	TotalTrips          int64   `json:"total_trips"`
	TotalPromotions     int64   `json:"total_promotions"`
	PendingImageReports int64   `json:"pending_image_reports"`
	PendingStaging      int64   `json:"pending_staging"`
	AvgRating           float64 `json:"avg_rating"`
	TotalReviewRatings  int64   `json:"total_review_ratings"`
}

type TopDestination struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	ReviewCount int64   `json:"review_count"`
	Location    string  `json:"location"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type SubRegionStat struct {
	SubRegion string `json:"sub_region"`
	Count     int64  `json:"count"`
}

func (h *Handler) GetOverview(c *gin.Context) {
	var resp OverviewResponse

	h.DB.Table("destinations").Where("deleted_at IS NULL").Count(&resp.TotalDestinations)
	h.DB.Table("events").Where("deleted_at IS NULL").Count(&resp.TotalEvents)
	h.DB.Table("users").Where("deleted_at IS NULL").Count(&resp.TotalUsers)
	h.DB.Table("reviews").Where("deleted_at IS NULL").Count(&resp.TotalReviews)
	h.DB.Table("stories").Where("deleted_at IS NULL").Count(&resp.TotalStories)
	h.DB.Table("partners").Where("deleted_at IS NULL").Count(&resp.TotalPartners)
	h.DB.Table("hotels").Where("deleted_at IS NULL").Count(&resp.TotalHotels)
	h.DB.Table("restaurants").Where("deleted_at IS NULL").Count(&resp.TotalRestaurants)
	h.DB.Table("guides").Where("deleted_at IS NULL").Count(&resp.TotalGuides)
	h.DB.Table("souvenirs").Where("deleted_at IS NULL").Count(&resp.TotalSouvenirs)
	h.DB.Table("rentals").Where("deleted_at IS NULL").Count(&resp.TotalRentals)
	h.DB.Table("articles").Where("deleted_at IS NULL").Count(&resp.TotalArticles)
	h.DB.Table("trips").Where("deleted_at IS NULL").Count(&resp.TotalTrips)
	h.DB.Table("promotions").Where("deleted_at IS NULL").Count(&resp.TotalPromotions)
	h.DB.Table("image_reports").Where("status = 'pending' AND deleted_at IS NULL").Count(&resp.PendingImageReports)
	h.DB.Table("staging_destinations").Where("status = 'pending' AND deleted_at IS NULL").Count(&resp.PendingStaging)

	h.DB.Table("reviews").Where("deleted_at IS NULL").Select("COALESCE(AVG(rating), 0)").Scan(&resp.AvgRating)
	h.DB.Table("reviews").Where("deleted_at IS NULL AND rating > 0").Count(&resp.TotalReviewRatings)

	httpx.Success(c, http.StatusOK, "Analytics overview", resp, nil)
}

func (h *Handler) GetTopDestinations(c *gin.Context) {
	type row struct {
		Name        string  `json:"name"`
		Category    string  `json:"category"`
		Rating      float64 `json:"rating"`
		ReviewCount int64   `json:"review_count"`
		Location    string  `json:"location"`
	}

	var results []row
	h.DB.Raw(`
		SELECT
			name,
			COALESCE(category, '') as category,
			COALESCE(rating, 0) as rating,
			COALESCE(review_count, 0) as review_count,
			COALESCE(location, '') as location
		FROM destinations
		WHERE deleted_at IS NULL
		ORDER BY review_count DESC, rating DESC
		LIMIT 10
	`).Scan(&results)

	httpx.Success(c, http.StatusOK, "Top destinations", results, nil)
}

func (h *Handler) GetCategoryStats(c *gin.Context) {
	type row struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}

	var results []row
	h.DB.Raw(`
		SELECT
			COALESCE(category, 'Uncategorized') as category,
			COUNT(*) as count
		FROM destinations
		WHERE deleted_at IS NULL
		GROUP BY category
		ORDER BY count DESC
	`).Scan(&results)

	httpx.Success(c, http.StatusOK, "Category stats", results, nil)
}

func (h *Handler) GetSubRegionStats(c *gin.Context) {
	type row struct {
		SubRegion string `json:"sub_region"`
		Count     int64  `json:"count"`
	}

	var results []row
	h.DB.Raw(`
		SELECT
			COALESCE(sub_region, 'Unknown') as sub_region,
			COUNT(*) as count
		FROM destinations
		WHERE deleted_at IS NULL
		GROUP BY sub_region
		ORDER BY count DESC
	`).Scan(&results)

	httpx.Success(c, http.StatusOK, "Sub-region stats", results, nil)
}

func (h *Handler) GetRecentActivity(c *gin.Context) {
	type activity struct {
		Type      string `json:"type"`
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}

	var results []activity
	h.DB.Raw(`
		SELECT * FROM (
			SELECT 'destination' as type, id, name, created_at::text FROM destinations WHERE deleted_at IS NULL
			UNION ALL
			SELECT 'event' as type, id, title as name, created_at::text FROM events WHERE deleted_at IS NULL
			UNION ALL
			SELECT 'review' as type, id, destination_id as name, created_at::text FROM reviews WHERE deleted_at IS NULL
			UNION ALL
			SELECT 'user' as type, id, name, created_at::text FROM users WHERE deleted_at IS NULL
			UNION ALL
			SELECT 'story' as type, id, title as name, created_at::text FROM stories WHERE deleted_at IS NULL
		) combined
		ORDER BY created_at DESC
		LIMIT 20
	`).Scan(&results)

	httpx.Success(c, http.StatusOK, "Recent activity", results, nil)
}

func (h *Handler) GetReportSummary(c *gin.Context) {
	type report struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		Value       int64  `json:"value"`
		Unit         string `json:"unit"`
	}

	var results []report

	// Total destinations
	var destCount int64
	h.DB.Table("destinations").Where("deleted_at IS NULL").Count(&destCount)
	results = append(results, report{Name: "Total Destinations", Type: "Count", Description: "All active destinations in the system", Value: destCount, Unit: "destinations"})

	// Total events
	var eventCount int64
	h.DB.Table("events").Where("deleted_at IS NULL").Count(&eventCount)
	results = append(results, report{Name: "Total Events", Type: "Count", Description: "All active events", Value: eventCount, Unit: "events"})

	// Total users
	var userCount int64
	h.DB.Table("users").Where("deleted_at IS NULL").Count(&userCount)
	results = append(results, report{Name: "Total Users", Type: "Count", Description: "All registered users", Value: userCount, Unit: "users"})

	// Total reviews
	var reviewCount int64
	h.DB.Table("reviews").Where("deleted_at IS NULL").Count(&reviewCount)
	results = append(results, report{Name: "Total Reviews", Type: "Count", Description: "All user reviews submitted", Value: reviewCount, Unit: "reviews"})

	// Average rating
	var avgRating float64
	h.DB.Table("reviews").Where("deleted_at IS NULL").Select("COALESCE(AVG(rating), 0)").Scan(&avgRating)
	avgRatingInt := int64(avgRating * 100)
	results = append(results, report{Name: "Average Rating", Type: "Average", Description: "Average rating across all destinations", Value: avgRatingInt, Unit: "rating_x100"})

	// Total stories
	var storyCount int64
	h.DB.Table("stories").Where("deleted_at IS NULL").Count(&storyCount)
	results = append(results, report{Name: "Travel Stories", Type: "Count", Description: "User-submitted travel stories", Value: storyCount, Unit: "stories"})

	// Pending image reports
	var pendingReports int64
	h.DB.Table("image_reports").Where("status = 'pending' AND deleted_at IS NULL").Count(&pendingReports)
	results = append(results, report{Name: "Pending Image Reports", Type: "Count", Description: "Image reports awaiting moderation", Value: pendingReports, Unit: "reports"})

	// Pending staging items
	var pendingStaging int64
	h.DB.Table("staging_destinations").Where("status = 'pending' AND deleted_at IS NULL").Count(&pendingStaging)
	results = append(results, report{Name: "Pending Staging Items", Type: "Count", Description: "Scraped destinations awaiting review", Value: pendingStaging, Unit: "items"})

	// Top category
	type catRow struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	var topCats []catRow
	h.DB.Raw(`
		SELECT COALESCE(category, 'Uncategorized') as category, COUNT(*) as count
		FROM destinations WHERE deleted_at IS NULL
		GROUP BY category ORDER BY count DESC LIMIT 5
	`).Scan(&topCats)
	for _, cat := range topCats {
		results = append(results, report{Name: "Category: " + cat.Category, Type: "Category", Description: "Destinations in category", Value: cat.Count, Unit: "destinations"})
	}

	httpx.Success(c, http.StatusOK, "Report summary", results, nil)
}
