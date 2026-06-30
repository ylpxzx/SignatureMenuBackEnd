package home

import (
	"sort"

	"signature-menu-backend/internal/httpx"
	"signature-menu-backend/internal/middleware"
	"signature-menu-backend/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

type summaryResponse struct {
	Stats           statsResponse             `json:"stats"`
	FeaturedRecipes []recipeSummaryResponse   `json:"featured_recipes"`
	RecentRecipes   []recipeSummaryResponse   `json:"recent_recipes"`
	Ingredients     []store.IngredientSummary `json:"ingredients"`
}

type statsResponse struct {
	TotalRecipes     int `json:"total_recipes"`
	AvailableRecipes int `json:"available_recipes"`
	TotalIngredients int `json:"total_ingredients"`
	TotalRecipeSteps int `json:"total_recipe_steps"`
}

type recipeSummaryResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Difficulty       int    `json:"difficulty"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	IngredientCount  int    `json:"ingredient_count"`
	StepCount        int    `json:"step_count"`
	IsAvailable      bool   `json:"is_available"`
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/home/summary", h.summary)
}

func (h *Handler) summary(c *gin.Context) {
	userID := middleware.UserID(c)
	recipes := h.store.ListRecipes(userID)
	ingredients := h.store.IngredientSummaries(userID)

	stats := statsResponse{
		TotalRecipes:     len(recipes),
		TotalIngredients: len(ingredients),
	}
	for _, item := range recipes {
		if item.IsAvailable {
			stats.AvailableRecipes++
		}
		stats.TotalRecipeSteps += len(item.Steps)
	}

	featured := append([]store.Recipe(nil), recipes...)
	sort.SliceStable(featured, func(i, j int) bool {
		if featured[i].Difficulty == featured[j].Difficulty {
			return featured[i].UpdatedAt.After(featured[j].UpdatedAt)
		}
		return featured[i].Difficulty > featured[j].Difficulty
	})

	httpx.OK(c, summaryResponse{
		Stats:           stats,
		FeaturedRecipes: summarizeRecipes(featured, 4),
		RecentRecipes:   summarizeRecipes(recipes, 5),
		Ingredients:     limitIngredients(ingredients, 8),
	})
}

func summarizeRecipes(items []store.Recipe, limit int) []recipeSummaryResponse {
	if len(items) < limit {
		limit = len(items)
	}
	response := make([]recipeSummaryResponse, 0, limit)
	for _, item := range items[:limit] {
		response = append(response, recipeSummaryResponse{
			ID:               item.ID,
			Name:             item.Name,
			Description:      item.Description,
			Difficulty:       item.Difficulty,
			EstimatedMinutes: item.EstimatedMinutes,
			IngredientCount:  len(item.Ingredients),
			StepCount:        len(item.Steps),
			IsAvailable:      item.IsAvailable,
		})
	}
	return response
}

func limitIngredients(items []store.IngredientSummary, limit int) []store.IngredientSummary {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}
