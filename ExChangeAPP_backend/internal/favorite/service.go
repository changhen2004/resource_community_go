package favorite

import (
	"context"
	"strconv"

	internalArticle "exchangeapp/internal/article"
	"exchangeapp/internal/asyncjob"
)

type Service struct {
	repo           *Repo
	articleService *internalArticle.Service
	publisher      asyncjob.Publisher
}

func NewService(repo *Repo, articleService *internalArticle.Service, publisher asyncjob.Publisher) *Service {
	if publisher == nil {
		publisher = asyncjob.NoopPublisher{}
	}
	return &Service{repo: repo, articleService: articleService, publisher: publisher}
}

func (s *Service) Create(articleID string, userID uint) (FavoriteActionResponse, error) {
	parsedArticleID, err := strconv.ParseUint(articleID, 10, 64)
	if err != nil {
		return FavoriteActionResponse{}, ErrArticleNotFound
	}

	exists, err := s.repo.ArticleExists(uint(parsedArticleID))
	if err != nil {
		return FavoriteActionResponse{}, err
	}
	if !exists {
		return FavoriteActionResponse{}, ErrArticleNotFound
	}

	hasFavorite, err := s.repo.HasFavorite(uint(parsedArticleID), userID)
	if err != nil {
		return FavoriteActionResponse{}, err
	}
	if hasFavorite {
		return FavoriteActionResponse{}, ErrAlreadyFavorited
	}

	count, err := s.repo.Create(uint(parsedArticleID), userID)
	if err != nil {
		return FavoriteActionResponse{}, err
	}
	s.repo.InvalidateArticleCaches(context.Background(), uint(parsedArticleID))
	if err := s.publisher.Publish(context.Background(), asyncjob.Job{
		Type: asyncjob.TypeFavoriteCreated,
		Payload: map[string]uint{
			"articleID": uint(parsedArticleID),
		},
	}); err != nil {
		if s.articleService != nil {
			if err := s.articleService.RecordFavoriteHeat(context.Background(), uint(parsedArticleID), true); err != nil {
				return FavoriteActionResponse{}, err
			}
		}
	}

	return FavoriteActionResponse{
		Message:       "article favorited successfully",
		FavoriteCount: count,
	}, nil
}

func (s *Service) Delete(articleID string, userID uint) (FavoriteActionResponse, error) {
	parsedArticleID, err := strconv.ParseUint(articleID, 10, 64)
	if err != nil {
		return FavoriteActionResponse{}, ErrFavoriteNotFound
	}

	hasFavorite, err := s.repo.HasFavorite(uint(parsedArticleID), userID)
	if err != nil {
		return FavoriteActionResponse{}, err
	}
	if !hasFavorite {
		return FavoriteActionResponse{}, ErrFavoriteNotFound
	}

	count, err := s.repo.Delete(uint(parsedArticleID), userID)
	if err != nil {
		return FavoriteActionResponse{}, err
	}
	s.repo.InvalidateArticleCaches(context.Background(), uint(parsedArticleID))
	if err := s.publisher.Publish(context.Background(), asyncjob.Job{
		Type: asyncjob.TypeFavoriteDeleted,
		Payload: map[string]uint{
			"articleID": uint(parsedArticleID),
		},
	}); err != nil {
		if s.articleService != nil {
			if err := s.articleService.RecordFavoriteHeat(context.Background(), uint(parsedArticleID), false); err != nil {
				return FavoriteActionResponse{}, err
			}
		}
	}

	return FavoriteActionResponse{
		Message:       "article unfavorited successfully",
		FavoriteCount: count,
	}, nil
}

func (s *Service) ListByUserID(userID uint) ([]FavoriteArticleResponse, error) {
	return s.repo.ListByUserID(userID)
}
