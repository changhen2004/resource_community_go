package comment

import (
	"context"
	"errors"
	"strconv"

	internalArticle "exchangeapp/internal/article"
	internalPoints "exchangeapp/internal/points"
	"gorm.io/gorm"
)

type Service struct {
	repo           *Repo
	articleService *internalArticle.Service
	pointsService  *internalPoints.Service
}

func NewService(repo *Repo, articleService *internalArticle.Service, pointsService *internalPoints.Service) *Service {
	return &Service{repo: repo, articleService: articleService, pointsService: pointsService}
}

func (s *Service) Create(ctx context.Context, articleID string, req CreateCommentRequest) (CommentResponse, error) {
	parsedArticleID, err := strconv.ParseUint(articleID, 10, 64)
	if err != nil {
		return CommentResponse{}, ErrArticleNotFound
	}

	exists, err := s.repo.ArticleExists(uint(parsedArticleID))
	if err != nil {
		return CommentResponse{}, err
	}
	if !exists {
		return CommentResponse{}, ErrArticleNotFound
	}

	comment := &Comment{
		ArticleID: uint(parsedArticleID),
		UserID:    userIDFromContext(ctx),
		Content:   req.Content,
	}
	if err := s.repo.Create(comment); err != nil {
		return CommentResponse{}, err
	}
	if s.pointsService != nil {
		if err := s.pointsService.AwardQualityInteraction(comment.UserID, comment.ID); err != nil {
			return CommentResponse{}, err
		}
	}
	if s.articleService != nil {
		if err := s.articleService.RecordCommentHeat(ctx, comment.ArticleID); err != nil {
			return CommentResponse{}, err
		}
	}

	author := CommentAuthorResponse{
		ID:       comment.UserID,
		Username: usernameFromContext(ctx),
	}
	if author.Username == "" {
		author, err = s.repo.FindAuthorByID(comment.UserID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return CommentResponse{}, err
		}
	}

	return CommentResponse{
		ID:        comment.ID,
		ArticleID: comment.ArticleID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		Author:    author,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}, nil
}

func (s *Service) List(articleID string) ([]CommentResponse, error) {
	parsedArticleID, err := strconv.ParseUint(articleID, 10, 64)
	if err != nil {
		return []CommentResponse{}, nil
	}

	return s.repo.ListByArticleID(uint(parsedArticleID))
}

func (s *Service) Delete(commentID string, userID uint) (DeleteCommentResponse, error) {
	parsedCommentID, err := strconv.ParseUint(commentID, 10, 64)
	if err != nil {
		return DeleteCommentResponse{}, ErrCommentNotFound
	}

	comment, err := s.repo.FindByID(uint(parsedCommentID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DeleteCommentResponse{}, ErrCommentNotFound
		}
		return DeleteCommentResponse{}, err
	}
	if comment.UserID != userID {
		return DeleteCommentResponse{}, ErrForbidden
	}
	if err := s.repo.Delete(comment); err != nil {
		return DeleteCommentResponse{}, err
	}
	if s.articleService != nil {
		if err := s.articleService.RevertCommentHeat(context.Background(), comment.ArticleID); err != nil {
			return DeleteCommentResponse{}, err
		}
	}

	return DeleteCommentResponse{Message: "comment deleted successfully"}, nil
}

func userIDFromContext(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}

	userID, ok := ctx.Value("userID").(uint)
	if !ok {
		return 0
	}
	return userID
}

func usernameFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	username, ok := ctx.Value("username").(string)
	if !ok {
		return ""
	}
	return username
}
