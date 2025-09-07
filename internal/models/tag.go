package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Tag タグモデル
type Tag struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"size:100;uniqueIndex:idx_tag_name;not null" json:"name" validate:"required,min=1,max=100"`
    Slug      string    `gorm:"size:100;uniqueIndex:idx_tag_slug;not null" json:"slug"`
    Color     string    `gorm:"size:7;default:#007bff;index:idx_tag_color" json:"color" validate:"omitempty,hexcolor"`
    Description string  `gorm:"size:500" json:"description" validate:"omitempty,max=500"`
    PostCount uint      `gorm:"default:0;index:idx_tag_post_count" json:"post_count"`
    IsActive  bool      `gorm:"default:true;index:idx_tag_active" json:"is_active"`
    CreatedAt time.Time `gorm:"autoCreateTime;index:idx_tag_created_at" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // リレーション
    Posts []Post `gorm:"many2many:post_tags;constraint:OnDelete:CASCADE" json:"posts,omitempty"`
}

// TableName テーブル名を明示的に指定
func (Tag) TableName() string {
    return "tags"
}

// TagForCreate タグ作成用構造体
type TagForCreate struct {
    Name        string `json:"name" validate:"required,min=1,max=100"`
    Color       string `json:"color,omitempty" validate:"omitempty,hexcolor"`
    Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// TagForUpdate タグ更新用構造体
type TagForUpdate struct {
    Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
    Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
    Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
    IsActive    *bool   `json:"is_active,omitempty"`
}

// TagResponse API レスポンス用構造体
type TagResponse struct {
    ID          uint   `json:"id"`
    Name        string `json:"name"`
    Slug        string `json:"slug"`
    Color       string `json:"color"`
    Description string `json:"description"`
    PostCount   uint   `json:"post_count"`
    IsActive    bool   `json:"is_active"`
}

// TagWithStats 統計情報付きタグ構造体
type TagWithStats struct {
    ID              uint      `json:"id"`
    Name            string    `json:"name"`
    Slug            string    `json:"slug"`
    Color           string    `json:"color"`
    Description     string    `json:"description"`
    PostCount       uint      `json:"post_count"`
    PublishedCount  int       `json:"published_count"`
    RecentPostCount int       `json:"recent_post_count"` // 過去30日間
    IsActive        bool      `json:"is_active"`
    CreatedAt       time.Time `json:"created_at"`
}

// TagPopularity タグ人気度構造体
type TagPopularity struct {
    ID        uint   `json:"id"`
    Name      string `json:"name"`
    Slug      string `json:"slug"`
    PostCount uint   `json:"post_count"`
    ViewCount int64  `json:"view_count"`  // 関連投稿の総閲覧数
    Rank      int    `json:"rank"`        // 人気順位
}

// BeforeCreate 作成前処理
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
    // スラッグ自動生成
    if t.Slug == "" {
        t.Slug = t.GenerateSlug()
    }
    
    // デフォルトカラー設定
    if t.Color == "" {
        t.Color = "#007bff"
    }
    
    return t.Validate()
}

// BeforeUpdate 更新前処理
func (t *Tag) BeforeUpdate(tx *gorm.DB) error {
    // 名前が変更された場合、スラッグも更新
    if tx.Statement.Changed("name") {
        t.Slug = t.GenerateSlug()
    }
    
    return nil
}

// GenerateSlug スラッグ生成
func (t *Tag) GenerateSlug() string {
    // 日本語・英語対応のスラッグ生成
    slug := strings.ToLower(t.Name)
    
    // 特殊文字を削除・置換
    reg := regexp.MustCompile(`[^\w\-]`)
    slug = reg.ReplaceAllString(slug, "-")
    
    // 連続するハイフンを単一に
    reg = regexp.MustCompile(`-+`)
    slug = reg.ReplaceAllString(slug, "-")
    
    // 前後のハイフンを削除
    slug = strings.Trim(slug, "-")
    
    // 空の場合はランダムな文字列
    if slug == "" {
        slug = fmt.Sprintf("tag-%d", time.Now().Unix())
    }
    
    return slug
}

// Validate バリデーション実行
func (t *Tag) Validate() error {
    return ValidateStruct(t)
}

// UpdatePostCount 投稿数を更新
func (t *Tag) UpdatePostCount(tx *gorm.DB) error {
    var count int64
    if err := tx.Table("post_tags").Where("tag_id = ?", t.ID).Count(&count).Error; err != nil {
        return err
    }
    
    return tx.Model(t).UpdateColumn("post_count", count).Error
}

// IsPopular 人気タグかどうか判定
func (t *Tag) IsPopular(threshold uint) bool {
    return t.PostCount >= threshold
}

// GetHexColor カラーコードを取得（#付き）
func (t *Tag) GetHexColor() string {
    if strings.HasPrefix(t.Color, "#") {
        return t.Color
    }
    return "#" + t.Color
}