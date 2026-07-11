package article

import (
	"time"
)

type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"`
	Preview string `json:"preview" binding:"required,max=500"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type ArticleResponse struct {
	ID        uint      `json:"id"`
	AuthorID  uint      `json:"authorId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Preview   string    `json:"preview"`
	Status    string    `json:"status"`
	ViewCount uint      `json:"viewCount"`
	LikeCount uint      `json:"likeCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LikeResponse struct {
	Likes int `json:"likes"`
}

type LikeActionResponse struct {
	Message string `json:"message"`
	Likes   int    `json:"likes"`
}

func toArticleResponse(article Article) ArticleResponse {
	return ArticleResponse{
		ID:        article.ID,
		AuthorID:  article.AuthorID,
		Title:     article.Title,
		Content:   article.Content,
		Preview:   article.Preview,
		Status:    article.Status,
		ViewCount: article.ViewCount,
		LikeCount: article.LikeCount,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}
}

func toArticleResponses(articles []Article) []ArticleResponse {
	responses := make([]ArticleResponse, 0, len(articles))
	for _, article := range articles {
		responses = append(responses, toArticleResponse(article))
	}
	return responses
}
