// internal/models/post.go (拡張版)
package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// PostStatus 投稿ステータス
type PostStatus string

const (
    PostStatusDraft     PostStatus = "draft"
    PostStatusPublished PostStatus = "published"
    PostStatusArchived  PostStatus = "archived"
)

// Post 投稿モデル
type Post struct {
    ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    uint       `gorm:"not null;index:idx_post_user_id" json:"user_id" validate:"required"`
    Title     string     `gorm:"size:255;not null;index:idx_post_title" json:"title" validate:"required,min=1,max=255"`
    Slug      string     `gorm:"size:255;uniqueIndex:idx_post_slug;not null" json:"slug"`
    Body      string     `gorm:"type:text;not null" json:"body" validate:"required,min=1"`
    Excerpt   string     `gorm:"size:500" json:"excerpt"`
    Status    PostStatus `gorm:"size:20;not null;default:draft;index:idx_post_status" json:"status" validate:"required,oneof=draft published archived"`
    ViewCount uint       `gorm:"default:0;index:idx_post_view_count" json:"view_count"`
    CreatedAt time.Time  `gorm:"autoCreateTime;index:idx_post_created_at" json:"created_at"`
    UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

    // リレーション
    User     User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
    Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty"`
    Tags     []Tag     `gorm:"many2many:post_tags;constraint:OnDelete:CASCADE" json:"tags,omitempty"`
}

// TableName テーブル名を明示的に指定
func (Post) TableName() string {
    return "posts"
}

// PostForCreate 投稿作成用構造体
type PostForCreate struct {
    UserID uint     `json:"user_id" validate:"required"`
    Title  string   `json:"title" validate:"required,min=1,max=255"`
    Body   string   `json:"body" validate:"required,min=1"`
    Status PostStatus `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
    TagIDs []uint   `json:"tag_ids"`
}

// PostForUpdate 投稿更新用構造体
type PostForUpdate struct {
    Title  *string     `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
    Body   *string     `json:"body,omitempty" validate:"omitempty,min=1"`
    Status *PostStatus `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
    TagIDs []uint      `json:"tag_ids"`
}

// PostResponse API レスポンス用構造体
type PostResponse struct {
    ID          uint         `json:"id"`
    Title       string       `json:"title"`
    Slug        string       `json:"slug"`
    Body        string       `json:"body"`
    Excerpt     string       `json:"excerpt"`
    Status      PostStatus   `json:"status"`
    ViewCount   uint         `json:"view_count"`
    CreatedAt   time.Time    `json:"created_at"`
    User        UserResponse `json:"user"`
    Tags        []Tag        `json:"tags,omitempty"`
    CommentCount int         `json:"comment_count,omitempty"`
}

// PostSummary 投稿サマリー用構造体
type PostSummary struct {
    ID           uint       `json:"id"`
    Title        string     `json:"title"`
    Slug         string     `json:"slug"`
    Excerpt      string     `json:"excerpt"`
    Status       PostStatus `json:"status"`
    ViewCount    uint       `json:"view_count"`
    CreatedAt    time.Time  `json:"created_at"`
    UserName     string     `json:"user_name"`
    CommentCount int        `json:"comment_count"`
    TagNames     []string   `json:"tag_names"`
}

// BeforeCreate 作成前処理（スラッグ生成）
func (p *Post) BeforeCreate(tx *gorm.DB) error {
    if p.Slug == "" {
        p.Slug = p.generateSlug()
    }
    if p.Excerpt == "" {
        p.Excerpt = p.generateExcerpt()
    }
    return nil
}

// BeforeUpdate 更新前処理
func (p *Post) BeforeUpdate(tx *gorm.DB) error {
    if p.Excerpt == "" {
        p.Excerpt = p.generateExcerpt()
    }
    return nil
}

// generateSlug スラッグ生成
func (p *Post) generateSlug() string {
    slug := strings.ToLower(p.Title)
    slug = strings.ReplaceAll(slug, " ", "-")
    // 実際には、より高度なスラッグ生成ロジックを実装
    return slug
}

// generateExcerpt 抜粋生成
func (p *Post) generateExcerpt() string {
    if len(p.Body) <= 200 {
        return p.Body
    }
    return p.Body[:200] + "..."
}

// Validate バリデーション実行
func (p *Post) Validate() error {
    return ValidateStruct(p)
}

// IsPublished 公開済みか確認
func (p *Post) IsPublished() bool {
    return p.Status == PostStatusPublished
}

// IncrementViewCount 閲覧数増加
func (p *Post) IncrementViewCount(tx *gorm.DB) error {
    return tx.Model(p).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}