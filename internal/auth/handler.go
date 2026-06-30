package auth

import (
	"errors"
	"net/http"
	"strings"

	"signature-menu-backend/internal/httpx"
	"signature-menu-backend/internal/middleware"
	"signature-menu-backend/internal/store"
	"signature-menu-backend/pkg/token"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	store  *store.Store
	tokens *token.Manager
}

func NewHandler(store *store.Store, tokens *token.Manager) *Handler {
	return &Handler{store: store, tokens: tokens}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup, protected *gin.RouterGroup) {
	api.POST("/auth/register", h.register)
	api.POST("/auth/login", h.login)
	protected.POST("/auth/logout", h.logout)
	protected.GET("/me", h.me)
}

func (h *Handler) register(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}
	if message := validateCredentials(req); message != "" {
		httpx.Error(c, http.StatusBadRequest, "invalid_input", message)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "password_hash_failed", "密码处理失败")
		return
	}

	user, err := h.store.CreateUser(req.Username, string(passwordHash), req.DisplayName)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			httpx.Error(c, http.StatusConflict, "username_taken", "这个账号已经注册")
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "create_user_failed", "注册失败")
		return
	}

	h.respondWithToken(c, http.StatusCreated, user)
}

func (h *Handler) login(c *gin.Context) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, http.StatusBadRequest, "bad_request", "请求格式不正确")
		return
	}

	user, err := h.store.FindUserByUsername(req.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		httpx.Error(c, http.StatusUnauthorized, "invalid_credentials", "账号或密码不正确")
		return
	}

	h.respondWithToken(c, http.StatusOK, user)
}

func (h *Handler) logout(c *gin.Context) {
	httpx.OK(c, gin.H{"message": "已退出登录"})
}

func (h *Handler) me(c *gin.Context) {
	user, err := h.store.GetUser(middleware.UserID(c))
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, "unauthorized", "请重新登录")
		return
	}
	httpx.OK(c, toUserResponse(user))
}

func (h *Handler) respondWithToken(c *gin.Context, status int, user store.User) {
	rawToken, err := h.tokens.Generate(user.ID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "token_failed", "登录凭证生成失败")
		return
	}
	c.JSON(status, gin.H{"data": authResponse{Token: rawToken, User: toUserResponse(user)}})
}

func validateCredentials(req credentialsRequest) string {
	username := strings.TrimSpace(req.Username)
	if len(username) < 3 {
		return "账号至少需要 3 个字符"
	}
	if len(req.Password) < 6 {
		return "密码至少需要 6 个字符"
	}
	return ""
}

func toUserResponse(user store.User) userResponse {
	return userResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
