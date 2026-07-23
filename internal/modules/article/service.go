package article

import (
	"errors"
	"time"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll(status string) ([]Article, error) {
	return s.Repo.FindAll(status)
}

func (s *Service) GetBySlug(slug string) (*Article, error) {
	return s.Repo.FindBySlug(slug)
}

func (s *Service) GetByID(externalID string) (*Article, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetByCategory(category string) ([]Article, error) {
	return s.Repo.FindByCategory(category)
}

func (s *Service) Search(query string) ([]Article, error) {
	return s.Repo.Search(query)
}

func (s *Service) Create(a *Article) error {
	return s.Repo.Create(a)
}

type UpdateArticleRequest struct {
	Title            *string  `json:"title"`
	TitleEn          *string  `json:"title_en"`
	Slug             *string  `json:"slug"`
	Excerpt          *string  `json:"excerpt"`
	ExcerptEn        *string  `json:"excerpt_en"`
	Content          *string  `json:"content"`
	ContentEn        *string  `json:"content_en"`
	CoverImage       *string  `json:"cover_image"`
	Category         *string  `json:"category"`
	Tags             *JSONArr `json:"tags"`
	Author           *string  `json:"author"`
	Status           *string  `json:"status"`
	PublishedAt      *string  `json:"published_at"`
	SeoTitle         *string  `json:"seo_title"`
	SeoTitleEn       *string  `json:"seo_title_en"`
	SeoDescription   *string  `json:"seo_description"`
	SeoDescriptionEn *string  `json:"seo_description_en"`
	SeoKeywords      *string  `json:"seo_keywords"`
	SeoKeywordsEn    *string  `json:"seo_keywords_en"`
	OgImage          *string  `json:"og_image"`
	ReadTimeMinutes  *int     `json:"read_time_minutes"`
}

func (s *Service) Update(externalID string, req UpdateArticleRequest) (*Article, error) {
	a, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("article not found")
	}

	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.TitleEn != nil {
		a.TitleEn = *req.TitleEn
	}
	if req.Slug != nil {
		a.Slug = *req.Slug
	}
	if req.Excerpt != nil {
		a.Excerpt = *req.Excerpt
	}
	if req.ExcerptEn != nil {
		a.ExcerptEn = *req.ExcerptEn
	}
	if req.Content != nil {
		a.Content = *req.Content
	}
	if req.ContentEn != nil {
		a.ContentEn = *req.ContentEn
	}
	if req.CoverImage != nil {
		a.CoverImage = *req.CoverImage
	}
	if req.Category != nil {
		a.Category = *req.Category
	}
	if req.Tags != nil {
		a.Tags = *req.Tags
	}
	if req.Author != nil {
		a.Author = *req.Author
	}
	if req.Status != nil {
		a.Status = *req.Status
		// Auto-set published_at when publishing for the first time
		if *req.Status == "published" && a.PublishedAt == nil {
			now := time.Now()
			a.PublishedAt = &now
		}
	}
	if req.PublishedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err == nil {
			a.PublishedAt = &t
		}
	}
	if req.SeoTitle != nil {
		a.SeoTitle = *req.SeoTitle
	}
	if req.SeoTitleEn != nil {
		a.SeoTitleEn = *req.SeoTitleEn
	}
	if req.SeoDescription != nil {
		a.SeoDescription = *req.SeoDescription
	}
	if req.SeoDescriptionEn != nil {
		a.SeoDescriptionEn = *req.SeoDescriptionEn
	}
	if req.SeoKeywords != nil {
		a.SeoKeywords = *req.SeoKeywords
	}
	if req.SeoKeywordsEn != nil {
		a.SeoKeywordsEn = *req.SeoKeywordsEn
	}
	if req.OgImage != nil {
		a.OgImage = *req.OgImage
	}
	if req.ReadTimeMinutes != nil {
		a.ReadTimeMinutes = *req.ReadTimeMinutes
	}

	if err := s.Repo.Update(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}
