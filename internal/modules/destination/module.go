package destination

import (
	"pleco-api/internal/ai"
	"pleco-api/internal/cache"
	"pleco-api/internal/search"

	"gorm.io/gorm"
)

type Module struct {
	Repository         Repository
	Service            *Service
	Handler            *Handler
	ContentGenHandler  *ContentGenHandler
	LocationHandler    *LocationHandler
}

func BuildModule(db *gorm.DB, cacheStore cache.Store, contentAIService *ai.Service, searchClient *search.Client) *Module {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service, cacheStore)

	contentRepo := NewContentGenRepository(db)
	contentSvc := NewContentGenService(repository, contentRepo, contentAIService, searchClient)
	contentHandler := NewContentGenHandler(contentSvc)

	locationHandler := NewLocationHandler(handler, service, repository)

	return &Module{
		Repository:        repository,
		Service:           service,
		Handler:           handler,
		ContentGenHandler: contentHandler,
		LocationHandler:   locationHandler,
	}
}
