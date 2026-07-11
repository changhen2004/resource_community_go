package article

import (
	"context"
	"encoding/json"
	"fmt"
)

const CacheKeyPrefix = "articles:list:"

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateArticleRequest) (ArticleResponse, error) {
	article := &Article{
		Title:   req.Title,
		Content: req.Content,
		Preview: req.Preview,
		Tags:    joinTags(req.Tags),
		Status:  req.Status,
	}
	if article.Status == "" {
		article.Status = "draft"
	}

	if err := s.repo.Create(article); err != nil {
		return ArticleResponse{}, err
	}
	s.repo.DeleteArticlesCacheByPrefix(ctx, CacheKeyPrefix)

	return toArticleResponse(*article), nil
}

func (s *Service) List(ctx context.Context, query ListArticlesQuery) ([]ArticleResponse, error) {
	cacheKey := buildListCacheKey(query)
	cached, err := s.repo.GetArticlesCache(ctx, cacheKey)
	if err == nil {
		var responses []ArticleResponse
		if unmarshalErr := json.Unmarshal([]byte(cached), &responses); unmarshalErr == nil {
			return responses, nil
		}
	}

	articles, err := s.repo.List(query)
	if err != nil {
		return nil, err
	}

	responses := toArticleResponses(articles)
	if payload, marshalErr := json.Marshal(responses); marshalErr == nil {
		s.repo.SetArticlesCache(ctx, cacheKey, string(payload))
	}
	return responses, nil
}

func (s *Service) FindByID(id string) (ArticleResponse, error) {
	article, err := s.repo.FindByID(id)
	if err != nil {
		return ArticleResponse{}, err
	}

	return toArticleResponse(*article), nil
}

func (s *Service) Like(ctx context.Context, articleID string) (LikeActionResponse, error) {
	likes, err := s.repo.IncrementLike(ctx, articleID)
	if err != nil {
		return LikeActionResponse{}, err
	}

	return LikeActionResponse{
		Message: "Article liked successfully",
		Likes:   likes,
	}, nil
}

func (s *Service) GetLikes(ctx context.Context, articleID string) (LikeResponse, error) {
	likes, err := s.repo.GetLikeCount(ctx, articleID)
	if err != nil {
		return LikeResponse{}, err
	}
	return LikeResponse{Likes: likes}, nil
}

func buildListCacheKey(query ListArticlesQuery) string {
	return fmt.Sprintf(
		"%spage=%d:size=%d:sort=%s:keyword=%s:tag=%s",
		CacheKeyPrefix,
		query.Page,
		query.PageSize,
		query.Sort,
		query.Keyword,
		query.Tag,
	)
}
