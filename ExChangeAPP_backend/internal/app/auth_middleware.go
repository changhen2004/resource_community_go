package app

import (
	internalAuth "exchangeapp/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type responseEnvelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type errorData struct {
	ErrorCode string `json:"errorCode"`
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

func AuthMiddleware(authService *internalAuth.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			writeError(ctx, http.StatusUnauthorized, 10002, "Authorization header is missing", "UNAUTHORIZED")
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeError(ctx, http.StatusUnauthorized, 10002, "Authorization header must use Bearer scheme", "UNAUTHORIZED")
			ctx.Abort()
			return
		}

		claims, err := authService.ValidateAccessToken(strings.TrimSpace(parts[1]))
		if err != nil {
			writeError(ctx, http.StatusUnauthorized, 10002, err.Error(), "UNAUTHORIZED")
			ctx.Abort()
			return
		}
		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Next()
	}
}
