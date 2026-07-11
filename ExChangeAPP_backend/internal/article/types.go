package article

import (
	"strings"
	"time"
)

type CreateArticleRequest struct {
	Title   string   `json:"title" binding:"required,max=200"`
	Content string   `json:"content" binding:"required"`
	Preview string   `json:"preview" binding:"required,max=500"`
	Tags    []string `json:"tags"`
	Status  string   `json:"status" binding:"omitempty,oneof=draft published archived"`
}

type ListArticlesQuery struct {
	Page     int
	PageSize int
	Sort     string
	Keyword  string
	Tag      string
}

type ArticleResponse struct {
	ID        uint      `json:"id"`
	AuthorID  uint      `json:"authorId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Preview   string    `json:"preview"`
	Tags      []string  `json:"tags"`
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
		Tags:      splitTags(article.Tags),
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
