package article

import "gorm.io/gorm"

type Article struct {
	gorm.Model
	AuthorID  uint   `gorm:"index"`
	Title     string `binding:"required,max=200" gorm:"size:200;not null;index"`
	Content   string `binding:"required" gorm:"type:text;not null"`
	Preview   string `binding:"required,max=500" gorm:"size:500;not null"`
	Status    string `binding:"omitempty,oneof=draft published archived" gorm:"size:20;not null;default:draft;index"`
	ViewCount uint   `gorm:"not null;default:0"`
	LikeCount uint   `gorm:"not null;default:0"`
}
