package destination

import (
	"pleco-api/internal/cache"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, cacheStore cache.Store) *Module {
	repository := NewRepository(db)
	service := NewService(repository)
	handler := NewHandler(service, cacheStore)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
