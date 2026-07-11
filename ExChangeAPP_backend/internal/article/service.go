package article

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"exchangeapp/internal/cachekey"
	internalMedia "exchangeapp/internal/media"
	internalPoints "exchangeapp/internal/points"
	"gorm.io/gorm"
)

type articleDetailCachePayload struct {
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
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
}

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
	if err := s.repo.SetInitialHotScore(ctx, article.ID, initialHotScore(article.CreatedAt)); err != nil {
		return ArticleResponse{}, err
	}
	if s.pointsService != nil {
		if err := s.pointsService.AwardPublishResource(article.AuthorID, article.ID); err != nil {
			return ArticleResponse{}, err
		}
	}
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleListPrefix)
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleHotPrefix)

	return toArticleResponse(*article), nil
}

func (s *Service) List(ctx context.Context, query ListArticlesQuery) ([]ArticleResponse, error) {
	cacheKey := cachekey.ArticleListKey(query.Page, query.PageSize, query.Sort, query.Keyword, query.Tag)
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
		s.repo.SetArticlesCache(ctx, cacheKey, string(payload), cachekey.ArticleListTTL)
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
	ctx := context.Background()

	article, err := s.repo.IncrementView(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ArticleDetailResponse{}, ErrArticleNotFound
		}
		return ArticleDetailResponse{}, err
	}
	if err := s.repo.AddHotScore(ctx, article.ID, hotScoreView); err != nil {
		return ArticleDetailResponse{}, err
	}
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleHotPrefix)
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleListPrefix)

	cacheKey := cachekey.ArticleDetailKey(id)
	cached, err := s.repo.GetArticlesCache(ctx, cacheKey)
	if err == nil {
		var payload articleDetailCachePayload
		if unmarshalErr := json.Unmarshal([]byte(cached), &payload); unmarshalErr == nil {
			payload.Stats.ViewCount = article.ViewCount
			isUnlocked, unlockErr := s.resolveUnlockStatus(Article{
				Model:          gorm.Model{ID: article.ID},
				AuthorID:       payload.Author.ID,
				IsFree:         payload.IsFree,
				RequiredPoints: payload.RequiredPoints,
			}, currentUserID)
			if unlockErr != nil {
				return ArticleDetailResponse{}, unlockErr
			}
			if refreshedPayload, marshalErr := json.Marshal(payload); marshalErr == nil {
				s.repo.SetArticlesCache(ctx, cacheKey, string(refreshedPayload), cachekey.ArticleDetailTTL)
			}
			return payload.toResponse(isUnlocked), nil
		}
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

	response := toArticleDetailResponse(*article, author, isUnlocked)
	if payload, marshalErr := json.Marshal(newArticleDetailCachePayload(response)); marshalErr == nil {
		s.repo.SetArticlesCache(ctx, cacheKey, string(payload), cachekey.ArticleDetailTTL)
	}
	return response, nil
}

func (s *Service) Like(ctx context.Context, articleID string) (LikeActionResponse, error) {
	likes, err := s.repo.IncrementLike(ctx, articleID)
	if err != nil {
		return LikeActionResponse{}, err
	}
	parsedArticleID, parseErr := strconv.ParseUint(articleID, 10, 64)
	if parseErr == nil {
		if err := s.repo.AddHotScore(ctx, uint(parsedArticleID), hotScoreLike); err != nil {
			return LikeActionResponse{}, err
		}
	}
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleListPrefix)
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleHotPrefix)
	s.repo.DeleteArticleCacheKeys(ctx, cachekey.ArticleDetailKey(articleID))

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

func (s *Service) ListHot(ctx context.Context, limit int) ([]ArticleResponse, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	cacheKey := cachekey.ArticleHotKey(limit)
	cached, err := s.repo.GetArticlesCache(ctx, cacheKey)
	if err == nil {
		var responses []ArticleResponse
		if unmarshalErr := json.Unmarshal([]byte(cached), &responses); unmarshalErr == nil {
			return responses, nil
		}
	}

	if err := s.repo.SeedHotRanking(ctx, 200); err != nil {
		return nil, err
	}

	hotIDs, err := s.repo.GetHotArticleIDs(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	var articles []Article
	if len(hotIDs) > 0 {
		articles, err = s.repo.ListByIDs(hotIDs)
		if err != nil {
			return nil, err
		}
	} else {
		articles, err = s.repo.List(NewListArticlesQuery(1, limit, "hot", "", ""))
		if err != nil {
			return nil, err
		}
	}
	responses := toArticleResponses(articles)
	if payload, marshalErr := json.Marshal(responses); marshalErr == nil {
		s.repo.SetArticlesCache(ctx, cacheKey, string(payload), cachekey.ArticleHotTTL)
	}
	return responses, nil
}

func (s *Service) RecordCommentHeat(ctx context.Context, articleID uint) error {
	if err := s.repo.AddHotScore(ctx, articleID, hotScoreComment); err != nil {
		return err
	}
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleHotPrefix)
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleListPrefix)
	s.repo.DeleteArticleCacheKeys(ctx, cachekey.ArticleDetailKey(strconv.FormatUint(uint64(articleID), 10)))
	return nil
}

func (s *Service) RecordFavoriteHeat(ctx context.Context, articleID uint, increase bool) error {
	delta := float64(-hotScoreFavorite)
	if increase {
		delta = hotScoreFavorite
	}

	if err := s.repo.AddHotScore(ctx, articleID, delta); err != nil {
		return err
	}
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleHotPrefix)
	s.repo.DeleteArticlesCacheByPrefix(ctx, cachekey.ArticleListPrefix)
	s.repo.DeleteArticleCacheKeys(ctx, cachekey.ArticleDetailKey(strconv.FormatUint(uint64(articleID), 10)))
	return nil
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

func newArticleDetailCachePayload(detail ArticleDetailResponse) articleDetailCachePayload {
	return articleDetailCachePayload{
		ID:             detail.ID,
		Title:          detail.Title,
		Content:        detail.Content,
		Preview:        detail.Preview,
		CoverURL:       detail.CoverURL,
		ContentImages:  detail.ContentImages,
		Tags:           detail.Tags,
		Status:         detail.Status,
		Author:         detail.Author,
		Stats:          detail.Stats,
		IsFree:         detail.IsFree,
		RequiredPoints: detail.RequiredPoints,
		CreatedAt:      detail.CreatedAt,
		UpdatedAt:      detail.UpdatedAt,
	}
}

func (p articleDetailCachePayload) toResponse(isUnlocked bool) ArticleDetailResponse {
	return ArticleDetailResponse{
		ID:             p.ID,
		Title:          p.Title,
		Content:        p.Content,
		Preview:        p.Preview,
		CoverURL:       p.CoverURL,
		ContentImages:  p.ContentImages,
		Tags:           p.Tags,
		Status:         p.Status,
		Author:         p.Author,
		Stats:          p.Stats,
		IsFree:         p.IsFree,
		RequiredPoints: p.RequiredPoints,
		IsUnlocked:     isUnlocked,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
