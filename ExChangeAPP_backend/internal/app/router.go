package app

import (
	internalArticle "exchangeapp/internal/article"
	internalAuth "exchangeapp/internal/auth"
	internalComment "exchangeapp/internal/comment"
	internalExchange "exchangeapp/internal/exchange"
	internalFavorite "exchangeapp/internal/favorite"
	internalMedia "exchangeapp/internal/media"
	internalPoints "exchangeapp/internal/points"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB                   *gorm.DB
	RedisDB              *redis.Client
	UploadDir            string
	EnablePprof          bool
	SlowRequestThreshold time.Duration
}

func SetUpRouter(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ObservabilityMiddleware(deps.SlowRequestThreshold))
	r.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "ok",
		})
	})

	if deps.EnablePprof {
		pprof.Register(r)
	}

	authService := internalAuth.NewService(
		internalAuth.NewRepo(deps.DB, deps.RedisDB),
	)
	authHandler := internalAuth.NewHandler(
		authService,
	)
	pointsService := internalPoints.NewService(
		internalPoints.NewRepo(deps.DB, deps.RedisDB),
	)
	pointsHandler := internalPoints.NewHandler(pointsService)
	articleHandler := internalArticle.NewHandler(
		internalArticle.NewService(
			internalArticle.NewRepo(deps.DB, deps.RedisDB),
			pointsService,
		),
	)
	commentHandler := internalComment.NewHandler(
		internalComment.NewService(
			internalComment.NewRepo(deps.DB),
			pointsService,
		),
	)
	exchangeHandler := internalExchange.NewHandler(
		internalExchange.NewService(
			internalExchange.NewRepo(deps.DB),
		),
	)
	favoriteHandler := internalFavorite.NewHandler(
		internalFavorite.NewService(
			internalFavorite.NewRepo(deps.DB, deps.RedisDB),
		),
	)
	mediaHandler := internalMedia.NewHandler(
		internalMedia.NewService(deps.UploadDir),
	)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Static("/uploads", deps.UploadDir)

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", RateLimitMiddleware(deps.RedisDB, loginRateLimitRule), authHandler.Login)
		auth.POST("/register", RateLimitMiddleware(deps.RedisDB, registerRateLimitRule), authHandler.Register)
		auth.POST("/refresh", authHandler.Refresh)
	}

	publicAPI := r.Group("/api")
	{
		publicAPI.GET("/exchangeRates", exchangeHandler.GetExchangeRate)
		publicAPI.GET("/articles", articleHandler.GetArticles)
		publicAPI.GET("/articles/hot", articleHandler.GetHotArticles)
		publicAPI.GET("/articles/:id", articleHandler.GetArticleByID)
		publicAPI.GET("/articles/:id/like", articleHandler.GetArticleLikes)
		publicAPI.GET("/articles/:id/comments", commentHandler.GetComments)
	}

	protectedAPI := r.Group("/api")
	protectedAPI.Use(AuthMiddleware(authService))
	{
		protectedAPI.POST("/auth/logout", authHandler.Logout)
		protectedAPI.POST("/exchangeRates", RateLimitMiddleware(deps.RedisDB, publishRateLimitRule), exchangeHandler.CreateExchangeRate)
		protectedAPI.POST("/articles", RateLimitMiddleware(deps.RedisDB, publishRateLimitRule), articleHandler.CreateArticle)
		protectedAPI.POST("/articles/:id/like", articleHandler.LikeArticle)
		protectedAPI.POST("/articles/:id/unlock", pointsHandler.UnlockArticle)
		protectedAPI.POST("/articles/:id/comments", RateLimitMiddleware(deps.RedisDB, commentRateLimitRule), commentHandler.CreateComment)
		protectedAPI.DELETE("/comments/:id", commentHandler.DeleteComment)
		protectedAPI.POST("/articles/:id/favorite", favoriteHandler.CreateFavorite)
		protectedAPI.DELETE("/articles/:id/favorite", favoriteHandler.DeleteFavorite)
		protectedAPI.POST("/uploads/cover", mediaHandler.UploadCover)
		protectedAPI.POST("/uploads/content-images", mediaHandler.UploadContentImages)
		protectedAPI.GET("/me/favorites", favoriteHandler.ListMyFavorites)
		protectedAPI.GET("/me/points", pointsHandler.GetMyPoints)
		protectedAPI.GET("/me/points/records", pointsHandler.GetMyPointsRecords)
		protectedAPI.POST("/me/check-in", RateLimitMiddleware(deps.RedisDB, checkInRateLimitRule), pointsHandler.CheckIn)
		protectedAPI.POST("/me/points/redeem", pointsHandler.RedeemPrivilege)
	}
	return r
}
