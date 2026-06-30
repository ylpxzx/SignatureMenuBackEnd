package recipe

import (
	"errors"
	"net/http"
	"strings"

	"signature-menu-backend/internal/httpx"
	"signature-menu-backend/internal/middleware"
	"signature-menu-backend/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *store.Store
}

func NewHandler(store *store.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/recipes", h.list)
	group.POST("/recipes", h.create)
	group.GET("/recipes/:id", h.get)
	group.PUT("/recipes/:id", h.update)
	group.DELETE("/recipes/:id", h.delete)
	group.GET("/ingredients", h.ingredients)
}

func (h *Handler) list(c *gin.Context) {
	userID := middleware.UserID(c)
	recipes := h.store.ListRecipes(userID)

	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))
	available := strings.TrimSpace(c.Query("available"))

	response := make([]recipeResponse, 0, len(recipes))
	for _, item := range recipes {
		if keyword != "" && !strings.Contains(strings.ToLower(item.Name+" "+item.Description), keyword) {
			continue
		}
		if available == "true" && !item.IsAvailable {
			continue
		}
		response = append(response, toRecipeResponse(item))
	}
	httpx.OK(c, response)
}

func (h *Handler) create(c *gin.Context) {
	var req recipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}

	created, err := h.store.CreateRecipe(middleware.UserID(c), toMutation(req))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.Created(c, toRecipeResponse(created))
}

func (h *Handler) get(c *gin.Context) {
	item, err := h.store.GetRecipe(middleware.UserID(c), c.Param("id"))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.OK(c, toRecipeResponse(item))
}

func (h *Handler) update(c *gin.Context) {
	var req recipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}

	updated, err := h.store.UpdateRecipe(middleware.UserID(c), c.Param("id"), toMutation(req))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.OK(c, toRecipeResponse(updated))
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.store.DeleteRecipe(middleware.UserID(c), c.Param("id")); err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.NoContent(c)
}

func (h *Handler) ingredients(c *gin.Context) {
	httpx.OK(c, h.store.IngredientSummaries(middleware.UserID(c)))
}

func toMutation(req recipeRequest) store.RecipeMutation {
	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	ingredients := make([]store.IngredientMutation, 0, len(req.Ingredients))
	for _, item := range req.Ingredients {
		ingredients = append(ingredients, store.IngredientMutation{
			ID:     item.ID,
			Name:   item.Name,
			Amount: item.Amount,
			Unit:   item.Unit,
			Note:   item.Note,
		})
	}

	steps := make([]store.StepMutation, 0, len(req.Steps))
	for _, item := range req.Steps {
		steps = append(steps, store.StepMutation{
			ID:               item.ID,
			StepOrder:        item.StepOrder,
			Title:            item.Title,
			Description:      item.Description,
			EstimatedMinutes: item.EstimatedMinutes,
		})
	}

	return store.RecipeMutation{
		Name:             req.Name,
		Description:      req.Description,
		ServingCount:     req.ServingCount,
		EstimatedMinutes: req.EstimatedMinutes,
		Difficulty:       req.Difficulty,
		IsAvailable:      isAvailable,
		Ingredients:      ingredients,
		Steps:            steps,
	}
}

func toRecipeResponse(item store.Recipe) recipeResponse {
	ingredients := make([]ingredientResponse, 0, len(item.Ingredients))
	for _, ingredient := range item.Ingredients {
		ingredients = append(ingredients, ingredientResponse{
			ID:        ingredient.ID,
			Name:      ingredient.Name,
			Amount:    ingredient.Amount,
			Unit:      ingredient.Unit,
			Note:      ingredient.Note,
			CreatedAt: ingredient.CreatedAt,
			UpdatedAt: ingredient.UpdatedAt,
		})
	}

	steps := make([]stepResponse, 0, len(item.Steps))
	for _, step := range item.Steps {
		steps = append(steps, stepResponse{
			ID:               step.ID,
			StepOrder:        step.StepOrder,
			Title:            step.Title,
			Description:      step.Description,
			EstimatedMinutes: step.EstimatedMinutes,
			CreatedAt:        step.CreatedAt,
			UpdatedAt:        step.UpdatedAt,
		})
	}

	return recipeResponse{
		ID:               item.ID,
		Name:             item.Name,
		Description:      item.Description,
		ServingCount:     item.ServingCount,
		EstimatedMinutes: item.EstimatedMinutes,
		Difficulty:       item.Difficulty,
		IsAvailable:      item.IsAvailable,
		Ingredients:      ingredients,
		Steps:            steps,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func handleStoreError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.Error(c, http.StatusNotFound, "not_found", "没有找到对应数据")
	case errors.Is(err, store.ErrInvalidInput):
		httpx.Error(c, http.StatusBadRequest, "invalid_input", "请至少填写菜谱名称")
	default:
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
	}
}
