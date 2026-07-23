package appsetup

import (
	"log"
	"log/slog"
	"sync/atomic"
	"pleco-api/internal/ai"
	"pleco-api/internal/config"
	"pleco-api/internal/erroroptimizer"
	"pleco-api/internal/httpx"
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/auth"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"
	configmodule "pleco-api/internal/modules/config"
	"pleco-api/internal/modules/guide"
	"pleco-api/internal/modules/hotel"
	"pleco-api/internal/modules/imagereport"
	"pleco-api/internal/modules/partner"
	"pleco-api/internal/modules/promotion"
	"pleco-api/internal/modules/rental"
	"pleco-api/internal/modules/restaurant"
	"pleco-api/internal/modules/review"
	"pleco-api/internal/modules/role"
	"pleco-api/internal/modules/souvenir"
	"pleco-api/internal/modules/staging"
	"pleco-api/internal/modules/article"
	"pleco-api/internal/modules/story"
	"pleco-api/internal/modules/tourist"
	"pleco-api/internal/modules/trips"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/modules/sitemap"
	"pleco-api/internal/scraper"
	"pleco-api/internal/services"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var (
	scrapeAllRunning      atomic.Bool
	scrapeDestRunning     atomic.Bool
	scrapeEventsRunning   atomic.Bool
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService, rateStore middleware.RateLimitStore) error {
	api := router.Group("/")
	cacheStore := newCacheStore(cfg)
	aiService, err := ai.NewService(cfg.AI)
	if err != nil {
		return err
	}
	permissionModule := permission.BuildModule(db, cacheStore)
	auditModule := audit.BuildModule(db, aiService)
	roleModule := role.BuildModule(db, auditModule.Service, cacheStore)
	api.Use(middleware.SecurityAuditLogger(func(event middleware.SecurityAuditEvent) {
		auditModule.Service.SafeRecord(audit.RecordInput{
			ActorUserID: event.UserID,
			Action:      "security_http_event",
			Resource:    "http",
			Status:      "failed",
			Description: event.Description,
			IPAddress:   event.IPAddress,
			UserAgent:   event.UserAgent,
		})
	}))
	userModule := user.BuildModule(db, auditModule.Service, cacheStore)
	userModule.Handler.PermissionSvc = permissionModule.Service
	userModule.Handler.Cache = cacheStore
	classifier := &erroroptimizer.DefaultErrorClassifier{}
	errOptSvc := erroroptimizer.NewErrorOptimizerService(classifier, aiService, cacheStore, db, slog.Default())

	authModule := auth.BuildModule(db, cfg, userModule.Service, jwtService, auditModule.Service, permissionModule.Service, errOptSvc, cacheStore)

	tokenVersionSrc := accessTokenVersionAdapter{repo: userModule.Repository}
	auth.SetupRoutes(api, authModule.Handler, jwtService, rateStore, tokenVersionSrc, cfg)
	user.SetupRoutes(api, userModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	audit.SetupRoutes(api, auditModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	role.SetupRoutes(api, roleModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)

	destinationModule := destination.BuildModule(db, cacheStore)
	destination.SetupRoutes(api, destinationModule.Handler, jwtService)

	configModule := configmodule.BuildModule(db)
	configmodule.SetupRoutes(api, configModule.Handler)

	eventModule := event.BuildModule(db, cacheStore)
	event.SetupRoutes(api, eventModule.Handler, jwtService)

	hotelModule := hotel.BuildModule(db)
	hotel.SetupRoutes(api, hotelModule.Handler, jwtService)

	restaurantModule := restaurant.BuildModule(db)
	restaurant.SetupRoutes(api, restaurantModule.Handler, jwtService)

	partnerModule := partner.BuildModule(db)
	partner.SetupRoutes(api, partnerModule.Handler, jwtService)

	guideModule := guide.BuildModule(db)
	guide.SetupRoutes(api, guideModule.Handler, jwtService)

	souvenirModule := souvenir.BuildModule(db)
	souvenir.SetupRoutes(api, souvenirModule.Handler, jwtService)

	rentalModule := rental.BuildModule(db)
	rental.SetupRoutes(api, rentalModule.Handler, jwtService)

	reviewModule := review.BuildModule(db)
	review.SetupRoutes(api, reviewModule.Handler, jwtService)

	storyModule := story.BuildModule(db)
	story.SetupRoutes(api, storyModule.Handler, jwtService)

	articleModule := article.BuildModule(db)
	article.SetupRoutes(api, articleModule.Handler, jwtService)

	promotionModule := promotion.BuildModule(db)
	promotion.SetupRoutes(api, promotionModule.Handler, jwtService)

	touristModule := tourist.BuildModule(aiService, destinationModule.Repository, eventModule.Repository, cacheStore)
	tourist.SetupRoutes(api, touristModule.Handler)

	tripsModule := trips.BuildModule(db)
	trips.SetupRoutes(api, tripsModule.Handler, jwtService)

	stagingModule := staging.BuildModule(db, aiService)
	stagingModule.RegisterRoutes(api)

	imageReportModule := imagereport.BuildModule(db)
	imageReportModule.RegisterPublicRoutes(api.Group("destinations"))
	imageReportModule.RegisterAdminRoutes(api.Group("admin"))

	router.GET("/health", func(c *gin.Context) {
		httpx.Success(c, 200, "Health check ok", gin.H{"status": "ok"}, nil)
	})

	// Manual scraper trigger endpoints (for testing)
	router.GET("/admin/scrape", func(c *gin.Context) {
		if !scrapeAllRunning.CompareAndSwap(false, true) {
			httpx.ErrorWithCode(c, 409, "SCRAPE_RUNNING", "A scrape is already in progress")
			return
		}
		go func() {
			defer scrapeAllRunning.Store(false)
			log.Println("[scraper] manual run triggered via /admin/scrape")
			results := scraper.RunAll(db)
			for _, r := range results {
				log.Printf("[scraper] %s: dest(%d/%d) events(%d/%d) errors(%d)",
					r.Source, r.DestinationsInserted, r.DestinationsUpdated,
					r.EventsInserted, r.EventsUpdated, len(r.Errors))
			}
			log.Println("[scraper] manual run complete")
		}()
		httpx.Success(c, 202, "Scrape started in background", nil, nil)
	})
	router.GET("/admin/scrape/destinations", func(c *gin.Context) {
		if !scrapeDestRunning.CompareAndSwap(false, true) {
			httpx.ErrorWithCode(c, 409, "SCRAPE_RUNNING", "A destination scrape is already in progress")
			return
		}
		go func() {
			defer scrapeDestRunning.Store(false)
			log.Println("[scraper] manual destinations run triggered")
			scraper.RunDestinationsOnly(db)
			log.Println("[scraper] manual destinations run complete")
		}()
		httpx.Success(c, 202, "Destination scrape started in background", nil, nil)
	})
	router.GET("/admin/scrape/events", func(c *gin.Context) {
		if !scrapeEventsRunning.CompareAndSwap(false, true) {
			httpx.ErrorWithCode(c, 409, "SCRAPE_RUNNING", "An event scrape is already in progress")
			return
		}
		go func() {
			defer scrapeEventsRunning.Store(false)
			log.Println("[scraper] manual events run triggered")
			scraper.RunEventsOnly(db)
			log.Println("[scraper] manual events run complete")
		}()
		httpx.Success(c, 202, "Event scrape started in background", nil, nil)
	})

	// Sitemap
	sitemapHandler := sitemap.NewHandler(db, cfg.Email.FrontendURL)
	router.GET("/sitemap.xml", sitemapHandler.GetSitemap)

	router.GET("/health/live", func(c *gin.Context) {
		httpx.Success(c, 200, "Service is live", gin.H{"status": "ok"}, nil)
	})

	router.GET("/health/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			httpx.Error(c, 503, "Database connection error")
			return
		}

		if err := sqlDB.Ping(); err != nil {
			httpx.Error(c, 503, "Database ping failed")
			return
		}

		httpx.Success(c, 200, "Service is ready", gin.H{"status": "ok"}, nil)
	})
	return nil
}

func BuildRouter(db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService, rateStore middleware.RateLimitStore) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(middleware.BodySizeLimit(cfg.RequestBodyLimitBytes))
	router.Use(middleware.RequestID())
	router.Use(middleware.StructuredLogger())
	router.Use(middleware.RecoveryLogger())
	router.Use(middleware.SecurityHeaders())
	if err := RegisterRoutes(router, db, cfg, jwtService, rateStore); err != nil {
		return nil, err
	}
	return router, nil
}
