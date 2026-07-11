package worker

import (
	"context"
	"errors"

	internalArticle "exchangeapp/internal/article"
	"exchangeapp/internal/asyncjob"
	internalPoints "exchangeapp/internal/points"
)

type Processor struct {
	articleService *internalArticle.Service
	pointsService  *internalPoints.Service
}

func NewProcessor(articleService *internalArticle.Service, pointsService *internalPoints.Service) *Processor {
	return &Processor{
		articleService: articleService,
		pointsService:  pointsService,
	}
}

func (p *Processor) Handle(ctx context.Context, job asyncjob.Job) error {
	switch job.Type {
	case asyncjob.TypeArticlePublished:
		return p.handleArticlePublished(ctx, job.Payload)
	case asyncjob.TypeArticleViewed:
		return p.handleArticleViewed(ctx, job.Payload)
	case asyncjob.TypeArticleLiked:
		return p.handleArticleLiked(ctx, job.Payload)
	case asyncjob.TypeCommentCreated:
		return p.handleCommentCreated(ctx, job.Payload)
	case asyncjob.TypeCommentDeleted:
		return p.handleCommentDeleted(ctx, job.Payload)
	case asyncjob.TypeFavoriteCreated:
		return p.handleFavoriteCreated(ctx, job.Payload)
	case asyncjob.TypeFavoriteDeleted:
		return p.handleFavoriteDeleted(ctx, job.Payload)
	default:
		return errors.New("unsupported async job type")
	}
}

func (p *Processor) handleArticlePublished(ctx context.Context, payload map[string]uint) error {
	if p.articleService != nil {
		if err := p.articleService.SetInitialHeat(ctx, payload["articleID"]); err != nil {
			return err
		}
	}
	if p.pointsService != nil {
		return p.pointsService.AwardPublishResource(payload["userID"], payload["articleID"])
	}
	return nil
}

func (p *Processor) handleArticleViewed(ctx context.Context, payload map[string]uint) error {
	if p.articleService == nil {
		return nil
	}
	return p.articleService.RecordView(ctx, payload["articleID"])
}

func (p *Processor) handleArticleLiked(ctx context.Context, payload map[string]uint) error {
	if p.articleService == nil {
		return nil
	}
	return p.articleService.ApplyLike(ctx, payload["articleID"])
}

func (p *Processor) handleCommentCreated(ctx context.Context, payload map[string]uint) error {
	if p.pointsService != nil {
		if err := p.pointsService.AwardQualityInteraction(payload["userID"], payload["commentID"]); err != nil {
			return err
		}
	}
	if p.articleService != nil {
		return p.articleService.RecordCommentHeat(ctx, payload["articleID"])
	}
	return nil
}

func (p *Processor) handleCommentDeleted(ctx context.Context, payload map[string]uint) error {
	if p.articleService == nil {
		return nil
	}
	return p.articleService.RevertCommentHeat(ctx, payload["articleID"])
}

func (p *Processor) handleFavoriteCreated(ctx context.Context, payload map[string]uint) error {
	if p.articleService == nil {
		return nil
	}
	return p.articleService.RecordFavoriteHeat(ctx, payload["articleID"], true)
}

func (p *Processor) handleFavoriteDeleted(ctx context.Context, payload map[string]uint) error {
	if p.articleService == nil {
		return nil
	}
	return p.articleService.RecordFavoriteHeat(ctx, payload["articleID"], false)
}
