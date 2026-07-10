package controllers

import (
	"encoding/json"
	"errors"
	"exchangeapp/model"
	"net/http"
	"time"

	"exchangeapp/global"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var CacheKey = "articles"

func CreateArticle(ctx *gin.Context) {
	var article model.Article

	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := global.DB.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 清除缓存
	if err := global.RedisDB.Del(ctx, CacheKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, article)
}

func GetArticles(ctx *gin.Context) {
	CacheData, err := global.RedisDB.Get(ctx, CacheKey).Result()
	if err == redis.Nil {

		var articles []model.Article

		if err := global.DB.Find(&articles).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "No articles found"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		articlesJSON, err := json.Marshal(articles)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := global.RedisDB.Set(ctx, CacheKey, articlesJSON, 10*time.Minute).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, articles)

	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		var articles []model.Article
		if err := json.Unmarshal([]byte(CacheData), &articles); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, articles)
		return
	}
}

func GetArticleByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article model.Article

	if err := global.DB.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, article)
}
