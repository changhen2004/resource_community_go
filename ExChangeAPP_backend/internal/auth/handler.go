package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type responseEnvelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type errorData struct {
	ErrorCode string `json:"errorCode"`
}

func writeSuccess(ctx *gin.Context, status int, data interface{}) {
	ctx.JSON(status, responseEnvelope{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func writeError(ctx *gin.Context, status int, code int, message, errorCode string) {
	ctx.JSON(status, responseEnvelope{
		Code:    code,
		Message: message,
		Data: errorData{
			ErrorCode: errorCode,
		},
	})
}

func (h *Handler) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, 10001, err.Error(), "INVALID_REQUEST")
		return
	}

	resp, err := h.service.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUsernameExists):
			writeError(ctx, http.StatusConflict, 10004, err.Error(), "USERNAME_ALREADY_EXISTS")
		default:
			writeError(ctx, http.StatusInternalServerError, 10005, err.Error(), "INTERNAL_ERROR")
		}
		return
	}

	writeSuccess(ctx, http.StatusOK, resp)
}

func (h *Handler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, 10001, err.Error(), "INVALID_REQUEST")
		return
	}

	resp, err := h.service.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			writeError(ctx, http.StatusUnauthorized, 10002, err.Error(), "USER_NOT_FOUND")
		case errors.Is(err, ErrInvalidPassword):
			writeError(ctx, http.StatusUnauthorized, 10002, "Invalid password", "INVALID_PASSWORD")
		default:
			writeError(ctx, http.StatusInternalServerError, 10005, err.Error(), "INTERNAL_ERROR")
		}
		return
	}

	writeSuccess(ctx, http.StatusOK, resp)
}

func (h *Handler) Refresh(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, http.StatusBadRequest, 10001, err.Error(), "INVALID_REQUEST")
		return
	}

	resp, err := h.service.Refresh(req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken):
			writeError(ctx, http.StatusUnauthorized, 10002, err.Error(), "INVALID_REFRESH_TOKEN")
		case errors.Is(err, ErrUserNotFound):
			writeError(ctx, http.StatusUnauthorized, 10002, err.Error(), "USER_NOT_FOUND")
		default:
			writeError(ctx, http.StatusInternalServerError, 10005, err.Error(), "INTERNAL_ERROR")
		}
		return
	}

	writeSuccess(ctx, http.StatusOK, resp)
}

func (h *Handler) Logout(ctx *gin.Context) {
	if err := h.service.Logout(ctx.GetUint("userID")); err != nil {
		writeError(ctx, http.StatusInternalServerError, 10005, err.Error(), "INTERNAL_ERROR")
		return
	}

	writeSuccess(ctx, http.StatusOK, gin.H{})
}
