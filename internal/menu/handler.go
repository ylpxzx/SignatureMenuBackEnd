package menu

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
	group.GET("/menus", h.list)
	group.POST("/menus", h.create)
	group.GET("/menus/:id", h.get)
	group.PUT("/menus/:id", h.update)
	group.PATCH("/menus/:id/status", h.updateStatus)
	group.DELETE("/menus/:id", h.delete)
}

func (h *Handler) list(c *gin.Context) {
	menus := h.store.ListMenus(middleware.UserID(c))
	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))
	status := strings.TrimSpace(c.Query("status"))

	response := make([]menuResponse, 0, len(menus))
	for _, item := range menus {
		if status != "" && item.Status != status {
			continue
		}
		if keyword != "" && !menuMatchesKeyword(item, keyword) {
			continue
		}
		response = append(response, toMenuResponse(item))
	}
	httpx.OK(c, response)
}

func (h *Handler) create(c *gin.Context) {
	var req menuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}

	created, err := h.store.CreateMenu(middleware.UserID(c), toMutation(req))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.Created(c, toMenuResponse(created))
}

func (h *Handler) get(c *gin.Context) {
	item, err := h.store.GetMenu(middleware.UserID(c), c.Param("id"))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.OK(c, toMenuResponse(item))
}

func (h *Handler) update(c *gin.Context) {
	var req menuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}

	updated, err := h.store.UpdateMenu(middleware.UserID(c), c.Param("id"), toMutation(req))
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.OK(c, toMenuResponse(updated))
}

func (h *Handler) updateStatus(c *gin.Context) {
	var req menuStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}

	updated, err := h.store.UpdateMenuStatus(middleware.UserID(c), c.Param("id"), req.Status)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.OK(c, toMenuResponse(updated))
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.store.DeleteMenu(middleware.UserID(c), c.Param("id")); err != nil {
		handleStoreError(c, err)
		return
	}
	httpx.NoContent(c)
}

func toMutation(req menuRequest) store.MenuMutation {
	dishes := make([]store.MenuDishMutation, 0, len(req.Dishes))
	for _, item := range req.Dishes {
		dishes = append(dishes, store.MenuDishMutation{
			RecipeID: item.RecipeID,
			Name:     item.Name,
			Count:    item.Count,
		})
	}

	return store.MenuMutation{
		Title:      req.Title,
		Note:       req.Note,
		DateKey:    req.DateKey,
		Time:       req.Time,
		Status:     req.Status,
		DinerCount: req.DinerCount,
		RecipeIDs:  req.RecipeIDs,
		Dishes:     dishes,
	}
}

func toMenuResponse(item store.MenuRecord) menuResponse {
	dishes := make([]dishResponse, 0, len(item.Dishes))
	for _, dish := range item.Dishes {
		dishes = append(dishes, dishResponse{
			RecipeID: dish.RecipeID,
			Name:     dish.Name,
			Count:    dish.Count,
		})
	}

	return menuResponse{
		ID:         item.ID,
		Title:      item.Title,
		Note:       item.Note,
		DateKey:    item.DateKey,
		DateLabel:  item.DateLabel,
		Weekday:    item.Weekday,
		Time:       item.Time,
		Status:     item.Status,
		DinerCount: item.DinerCount,
		RecipeIDs:  item.RecipeIDs,
		Dishes:     dishes,
		Tone:       item.Tone,
		CreatedAt:  item.CreatedAt,
		UpdatedAt:  item.UpdatedAt,
	}
}

func menuMatchesKeyword(item store.MenuRecord, keyword string) bool {
	content := strings.ToLower(item.Title + " " + item.Note)
	if strings.Contains(content, keyword) {
		return true
	}
	for _, dish := range item.Dishes {
		if strings.Contains(strings.ToLower(dish.Name), keyword) {
			return true
		}
	}
	return false
}

func handleStoreError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httpx.Error(c, http.StatusNotFound, "not_found", "没有找到对应菜单")
	case errors.Is(err, store.ErrInvalidInput):
		httpx.Error(c, http.StatusBadRequest, "invalid_input", "请填写菜单名、出餐日期和出餐时间")
	default:
		httpx.Error(c, http.StatusInternalServerError, "internal_error", "服务暂时不可用")
	}
}
