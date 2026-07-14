package config

import (
	"github.com/gin-gonic/gin"
	"pleco-api/internal/httpx"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

type SubRegion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Quote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

var categories = []Category{
	{ID: "hidden-gem", Name: "Hidden Gems", Icon: "Sparkles", Description: "Unexplored, pristine secret wonders"},
	{ID: "nature", Name: "Nature Escapes", Icon: "Leaf", Description: "Verdant forests, mountains, and parks"},
	{ID: "culinary", Name: "Culinary Legends", Icon: "Utensils", Description: "Rich sweet-savory traditional tastes"},
	{ID: "heritage", Name: "Heritage & Culture", Icon: "Castle", Description: "Ancient empires and royal palaces"},
	{ID: "adventure", Name: "Adventure", Icon: "Compass", Description: "Thrilling volcanic offroads and caves"},
	{ID: "beach", Name: "Beaches & Sunsets", Icon: "Sun", Description: "Vast golden sand cliffside coastlines"},
	{ID: "family", Name: "Family Friendly", Icon: "Users", Description: "Amusements and cultural experiences"},
	{ID: "weekend", Name: "Weekend Ideas", Icon: "CalendarDays", Description: "Short-trip custom curated escapes"},
}

var subRegions = []SubRegion{
	{ID: "yogyakarta", Name: "Yogyakarta City", Description: "The cultural heart of the Sultanate"},
	{ID: "sleman", Name: "Sleman", Description: "The majestic highlands of Mount Merapi"},
	{ID: "bantul", Name: "Bantul", Description: "Dramatic southern beaches and pine forests"},
	{ID: "kulonprogo", Name: "Kulon Progo", Description: "Breathtaking hills, waterfalls, and tea plantations"},
	{ID: "gunungkidul", Name: "Gunungkidul", Description: "Rugged limestone cliffs, caves, and pristine white-sand beaches"},
}

var quotes = []Quote{
	{Text: "Jogja is made of comfortable homes, warm street food, and unforgettable memories.", Author: "Anies Baswedan"},
	{Text: "Every corner of Yogyakarta has its own story, whispering legends of empires past.", Author: "Javanese Proverb"},
	{Text: "You can leave Yogyakarta, but its soul will remain a part of you forever.", Author: "Traditional Song"},
}

func (h *Handler) GetCategories(c *gin.Context) {
	httpx.Success(c, 200, "Categories fetched", categories, nil)
}

func (h *Handler) GetSubRegions(c *gin.Context) {
	httpx.Success(c, 200, "Sub-regions fetched", subRegions, nil)
}

func (h *Handler) GetQuotes(c *gin.Context) {
	httpx.Success(c, 200, "Quotes fetched", quotes, nil)
}
