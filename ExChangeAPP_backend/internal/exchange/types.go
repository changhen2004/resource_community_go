package exchange

import (
	"time"
)

type CreateExchangeRateRequest struct {
	FromCurrency string  `json:"fromCurrency" binding:"required"`
	ToCurrency   string  `json:"toCurrency" binding:"required"`
	Rate         float64 `json:"rate" binding:"required,gt=0"`
}

type ExchangeRateResponse struct {
	ID           uint      `json:"id"`
	FromCurrency string    `json:"fromCurrency"`
	ToCurrency   string    `json:"toCurrency"`
	Rate         float64   `json:"rate"`
	Date         time.Time `json:"date"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func toExchangeRateResponse(rate ExchangeRate) ExchangeRateResponse {
	return ExchangeRateResponse{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		Date:         rate.Date,
	}
}

func toExchangeRateResponses(rates []ExchangeRate) []ExchangeRateResponse {
	responses := make([]ExchangeRateResponse, 0, len(rates))
	for _, rate := range rates {
		responses = append(responses, toExchangeRateResponse(rate))
	}
	return responses
}
