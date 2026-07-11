package asyncjob

type Type string

const (
	TypeArticlePublished Type = "article.published"
	TypeArticleViewed    Type = "article.viewed"
	TypeArticleLiked     Type = "article.liked"
	TypeCommentCreated   Type = "comment.created"
	TypeCommentDeleted   Type = "comment.deleted"
	TypeFavoriteCreated  Type = "favorite.created"
	TypeFavoriteDeleted  Type = "favorite.deleted"
)

type Job struct {
	Type    Type            `json:"type"`
	Payload map[string]uint `json:"payload"`
}
