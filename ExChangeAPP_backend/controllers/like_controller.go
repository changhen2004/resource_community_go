package controllers

import (
	"net/http"

	"exchangeapp/global"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func LikeArticle(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":like"

	if err := global.RedisDB.Incr(ctx, likeKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Article liked successfully"})

}

func GetArticleLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likeKey := "article:" + articleID + ":like"

	likeCount, err := global.RedisDB.Get(ctx, likeKey).Result()

	if err == redis.Nil {
		likeCount = "0"
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"like_count": likeCount})
}
