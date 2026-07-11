package article

import (
	"strings"
	"time"
)

type CreateArticleRequest struct {
	Title          string   `json:"title" binding:"required,max=200"`
	Content        string   `json:"content" binding:"required"`
	Preview        string   `json:"preview" binding:"required,max=500"`
	CoverURL       string   `json:"coverUrl"`
	ContentImages  []string `json:"contentImages"`
	Tags           []string `json:"tags"`
	Status         string   `json:"status" binding:"omitempty,oneof=draft published archived"`
	IsFree         *bool    `json:"isFree"`
	RequiredPoints uint     `json:"requiredPoints"`
}

type ListArticlesQuery struct {
	Page     int
	PageSize int
	Sort     string
	Keyword  string
	Tag      string
}

type ArticleResponse struct {
	ID             uint      `json:"id"`
	AuthorID       uint      `json:"authorId"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Preview        string    `json:"preview"`
	CoverURL       string    `json:"coverUrl"`
	ContentImages  []string  `json:"contentImages"`
	Tags           []string  `json:"tags"`
	Status         string    `json:"status"`
	ViewCount      uint      `json:"viewCount"`
	LikeCount      uint      `json:"likeCount"`
	FavoriteCount  uint      `json:"favoriteCount"`
	IsFree         bool      `json:"isFree"`
	RequiredPoints uint      `json:"requiredPoints"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ArticleAuthorResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type ArticleStatsResponse struct {
	ViewCount     uint `json:"viewCount"`
	LikeCount     uint `json:"likeCount"`
	FavoriteCount uint `json:"favoriteCount"`
}

type ArticleDetailResponse struct {
	ID             uint                  `json:"id"`
	Title          string                `json:"title"`
	Content        string                `json:"content"`
	Preview        string                `json:"preview"`
	CoverURL       string                `json:"coverUrl"`
	ContentImages  []string              `json:"contentImages"`
	Tags           []string              `json:"tags"`
	Status         string                `json:"status"`
	Author         ArticleAuthorResponse `json:"author"`
	Stats          ArticleStatsResponse  `json:"stats"`
	IsFree         bool                  `json:"isFree"`
	RequiredPoints uint                  `json:"requiredPoints"`
	IsUnlocked     bool                  `json:"isUnlocked"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
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
		ID:             article.ID,
		AuthorID:       article.AuthorID,
		Title:          article.Title,
		Content:        article.Content,
		Preview:        article.Preview,
		CoverURL:       article.CoverURL,
		ContentImages:  splitContentImages(article.ContentImages),
		Tags:           splitTags(article.Tags),
		Status:         article.Status,
		ViewCount:      article.ViewCount,
		LikeCount:      article.LikeCount,
		FavoriteCount:  article.FavoriteCount,
		IsFree:         article.IsFree,
		RequiredPoints: article.RequiredPoints,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
	}
}

func toArticleResponses(articles []Article) []ArticleResponse {
	responses := make([]ArticleResponse, 0, len(articles))
	for _, article := range articles {
		responses = append(responses, toArticleResponse(article))
	}
	return responses
}

func NewListArticlesQuery(page, pageSize int, sort, keyword, tag string) ListArticlesQuery {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	sort = strings.ToLower(strings.TrimSpace(sort))
	if sort != "hot" {
		sort = "latest"
	}

	return ListArticlesQuery{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Keyword:  strings.TrimSpace(keyword),
		Tag:      strings.ToLower(strings.TrimSpace(tag)),
	}
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		cleaned := strings.ToLower(strings.TrimSpace(tag))
		if cleaned == "" {
			continue
		}
		if _, exists := seen[cleaned]; exists {
			continue
		}
		seen[cleaned] = struct{}{}
		normalized = append(normalized, cleaned)
	}

	return normalized
}

func joinTags(tags []string) string {
	return strings.Join(normalizeTags(tags), ",")
}

func splitTags(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return []string{}
	}
	return strings.Split(tags, ",")
}

func joinContentImages(urls []string) string {
	return strings.Join(normalizeContentImages(urls), ",")
}

func splitContentImages(urls string) []string {
	if strings.TrimSpace(urls) == "" {
		return []string{}
	}
	return strings.Split(urls, ",")
}

func normalizeContentImages(urls []string) []string {
	if len(urls) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(urls))
	for _, url := range urls {
		cleaned := strings.TrimSpace(url)
		if cleaned == "" {
			continue
		}
		normalized = append(normalized, cleaned)
	}
	return normalized
}

func toArticleDetailResponse(article Article, author ArticleAuthorResponse, isUnlocked bool) ArticleDetailResponse {
	return ArticleDetailResponse{
		ID:            article.ID,
		Title:         article.Title,
		Content:       article.Content,
		Preview:       article.Preview,
		CoverURL:      article.CoverURL,
		ContentImages: splitContentImages(article.ContentImages),
		Tags:          splitTags(article.Tags),
		Status:        article.Status,
		Author:        author,
		Stats: ArticleStatsResponse{
			ViewCount:     article.ViewCount,
			LikeCount:     article.LikeCount,
			FavoriteCount: article.FavoriteCount,
		},
		IsFree:         article.IsFree,
		RequiredPoints: article.RequiredPoints,
		IsUnlocked:     isUnlocked,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
	}
}
