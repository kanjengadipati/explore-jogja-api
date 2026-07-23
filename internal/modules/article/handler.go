package article

import (
	"strings"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// resolveLocale reads the Accept-Language header and returns "id" (default) or "en".
func resolveLocale(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "id"
	}
	lang = strings.ToLower(strings.TrimSpace(strings.Split(lang, ",")[0]))
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	return "id"
}

func (h *Handler) GetAll(c *gin.Context) {
	locale := resolveLocale(c)
	// Admin can request all statuses; public only sees published
	status := c.Query("status")
	if status == "" {
		status = "published"
	}

	articles, err := h.Service.GetAll(status)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	pag := httpx.ParsePagination(c)
	total := int64(len(articles))

	localized := make([]Article, len(articles))
	for i, a := range articles {
		localized[i] = a.Localize(locale)
	}

	start := pag.Offset
	if start > len(localized) {
		start = len(localized)
	}
	end := start + pag.Limit
	if end > len(localized) {
		end = len(localized)
	}
	paged := localized[start:end]

	meta := httpx.BuildPaginationMeta(total, pag.Page(), pag.Limit)
	httpx.Success(c, 200, "Articles fetched", paged, meta)
}

func (h *Handler) GetBySlug(c *gin.Context) {
	locale := resolveLocale(c)
	slug := c.Param("slug")
	a, err := h.Service.GetBySlug(slug)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Article not found")
		return
	}
	localized := a.Localize(locale)
	httpx.Success(c, 200, "Article fetched", localized, nil)
}

func (h *Handler) GetByID(c *gin.Context) {
	locale := resolveLocale(c)
	id := c.Param("id")
	a, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Article not found")
		return
	}
	localized := a.Localize(locale)
	httpx.Success(c, 200, "Article fetched", localized, nil)
}

func (h *Handler) GetByCategory(c *gin.Context) {
	locale := resolveLocale(c)
	category := c.Param("category")
	articles, err := h.Service.GetByCategory(category)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	localized := make([]Article, len(articles))
	for i, a := range articles {
		localized[i] = a.Localize(locale)
	}
	httpx.Success(c, 200, "Articles fetched", localized, nil)
}

func (h *Handler) Search(c *gin.Context) {
	locale := resolveLocale(c)
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}
	articles, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	localized := make([]Article, len(articles))
	for i, a := range articles {
		localized[i] = a.Localize(locale)
	}
	httpx.Success(c, 200, "Search results", localized, nil)
}

func (h *Handler) Create(c *gin.Context) {
	var a Article
	if err := c.ShouldBindJSON(&a); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	if err := h.Service.Create(&a); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 201, "Article created", a, nil)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	a, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Article updated", a, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Article deleted", nil, nil)
}
