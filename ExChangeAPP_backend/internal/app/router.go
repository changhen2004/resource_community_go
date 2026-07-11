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
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	DB      *gorm.DB
	RedisDB *redis.Client
}

func SetUpRouter(deps Dependencies) *gin.Engine {
	r := gin.Default()

	authHandler := internalAuth.NewHandler(
		internalAuth.NewService(
			internalAuth.NewRepo(deps.DB),
		),
	)
	pointsService := internalPoints.NewService(
		internalPoints.NewRepo(deps.DB),
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
			internalFavorite.NewRepo(deps.DB),
		),
	)
	mediaHandler := internalMedia.NewHandler(
		internalMedia.NewService(),
	)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Static("/uploads", "./uploads")

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}

	publicAPI := r.Group("/api")
	{
		publicAPI.GET("/exchangeRates", exchangeHandler.GetExchangeRate)
		publicAPI.GET("/articles", articleHandler.GetArticles)
		publicAPI.GET("/articles/:id", articleHandler.GetArticleByID)
		publicAPI.GET("/articles/:id/like", articleHandler.GetArticleLikes)
		publicAPI.GET("/articles/:id/comments", commentHandler.GetComments)
	}

	protectedAPI := r.Group("/api")
	protectedAPI.Use(AuthMiddleware())
	{
		protectedAPI.POST("/exchangeRates", exchangeHandler.CreateExchangeRate)
		protectedAPI.POST("/articles", articleHandler.CreateArticle)
		protectedAPI.POST("/articles/:id/like", articleHandler.LikeArticle)
		protectedAPI.POST("/articles/:id/unlock", pointsHandler.UnlockArticle)
		protectedAPI.POST("/articles/:id/comments", commentHandler.CreateComment)
		protectedAPI.DELETE("/comments/:id", commentHandler.DeleteComment)
		protectedAPI.POST("/articles/:id/favorite", favoriteHandler.CreateFavorite)
		protectedAPI.DELETE("/articles/:id/favorite", favoriteHandler.DeleteFavorite)
		protectedAPI.POST("/uploads/cover", mediaHandler.UploadCover)
		protectedAPI.POST("/uploads/content-images", mediaHandler.UploadContentImages)
		protectedAPI.GET("/me/favorites", favoriteHandler.ListMyFavorites)
		protectedAPI.GET("/me/points", pointsHandler.GetMyPoints)
		protectedAPI.GET("/me/points/records", pointsHandler.GetMyPointsRecords)
		protectedAPI.POST("/me/check-in", pointsHandler.CheckIn)
		protectedAPI.POST("/me/points/redeem", pointsHandler.RedeemPrivilege)
	}
	return r
}
