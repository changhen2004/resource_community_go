package router

import (
	"exchangeapp/controllers"
	"exchangeapp/middlewares"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetUpRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
	}

	publicAPI := r.Group("/api")
	{
		publicAPI.GET("/exchangeRates", controllers.GetExchangeRate)
		publicAPI.GET("/articles", controllers.GetArticles)
		publicAPI.GET("/articles/:id", controllers.GetArticleByID)
		publicAPI.GET("/articles/:id/like", controllers.GetArticleLikes)
	}

	protectedAPI := r.Group("/api")
	protectedAPI.Use(middlewares.AuthMiddleware())
	{
		protectedAPI.POST("/exchangeRates", controllers.CreateExchangeRate)
		protectedAPI.POST("/articles", controllers.CreateArticle)
		protectedAPI.POST("/articles/:id/like", controllers.LikeArticle)
	}
	return r
}
