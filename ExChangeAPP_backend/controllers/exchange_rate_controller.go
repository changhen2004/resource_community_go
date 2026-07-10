package controllers

import (
	"errors"
	"exchangeapp/model"
	"net/http"
	"time"

	"exchangeapp/global"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateExchangeRate(ctx *gin.Context) {
	var exchangeRate model.ExchangeRate
	if err := ctx.ShouldBindJSON(&exchangeRate); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exchangeRate.Date = time.Now()
	if err := global.DB.Create(&exchangeRate).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Exchange rate created successfully"})
}

func GetExchangeRate(ctx *gin.Context) {
	var exchangeRates []model.ExchangeRate

	if err := global.DB.Find(&exchangeRates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Exchange rate not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	ctx.JSON(http.StatusOK, exchangeRates)
}
