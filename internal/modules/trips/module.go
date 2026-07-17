package trips

import "gorm.io/gorm"

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB) *Module {
	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)
	return &Module{
		Repository: repo,
		Service:    svc,
		Handler:    handler,
	}
}
