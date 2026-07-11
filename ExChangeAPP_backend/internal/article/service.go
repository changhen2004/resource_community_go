package article

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	internalMedia "exchangeapp/internal/media"
	internalPoints "exchangeapp/internal/points"
	"gorm.io/gorm"
)

const CacheKeyPrefix = "articles:list:"

type Service struct {
	repo          *Repo
	pointsService *internalPoints.Service
}

func NewService(repo *Repo, pointsService *internalPoints.Service) *Service {
	return &Service{repo: repo, pointsService: pointsService}
}

func (s *Service) Create(ctx context.Context, req CreateArticleRequest) (ArticleResponse, error) {
	isFree := true
	if req.IsFree != nil {
		isFree = *req.IsFree
	}
	if req.RequiredPoints > 0 {
		isFree = false
	}

	article := &Article{
		AuthorID:       userIDFromContext(ctx),
		Title:          req.Title,
		Content:        req.Content,
		Preview:        req.Preview,
		CoverURL:       req.CoverURL,
		ContentImages:  joinContentImages(req.ContentImages),
		Tags:           joinTags(req.Tags),
		Status:         req.Status,
		IsFree:         isFree,
		RequiredPoints: req.RequiredPoints,
	}
	if article.Status == "" {
		article.Status = "draft"
	}
	if len(normalizeContentImages(req.ContentImages)) > internalMedia.ContentImageMaxCount {
		return ArticleResponse{}, ErrTooManyContentImages
	}

	if err := s.repo.Create(article); err != nil {
		return ArticleResponse{}, err
	}
	if s.pointsService != nil {
		if err := s.pointsService.AwardPublishResource(article.AuthorID, article.ID); err != nil {
			return ArticleResponse{}, err
		}
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

func (s *Service) GetDetail(id string, currentUserID uint) (ArticleDetailResponse, error) {
	article, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ArticleDetailResponse{}, ErrArticleNotFound
		}
		return ArticleDetailResponse{}, err
	}

	author, err := s.repo.FindAuthorByID(article.AuthorID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			author = ArticleAuthorResponse{}
		} else {
			return ArticleDetailResponse{}, err
		}
	}

	isUnlocked, err := s.resolveUnlockStatus(*article, currentUserID)
	if err != nil {
		return ArticleDetailResponse{}, err
	}

	return toArticleDetailResponse(*article, author, isUnlocked), nil
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

func (s *Service) resolveUnlockStatus(article Article, currentUserID uint) (bool, error) {
	if article.IsFree || article.RequiredPoints == 0 {
		return true, nil
	}
	if currentUserID == 0 {
		return false, nil
	}
	if article.AuthorID == currentUserID {
		return true, nil
	}
	return s.repo.HasUnlocked(article.ID, currentUserID)
}

func userIDFromContext(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}

	value := ctx.Value("userID")
	userID, ok := value.(uint)
	if !ok {
		return 0
	}
	return userID
}
