package article

import "errors"

var (
	ErrArticleNotFound      = errors.New("article not found")
	ErrTooManyContentImages = errors.New("too many content images")
)
