// internal/models/tag.go
package models

import "time"

// Tag タグモデル
type Tag struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // リレーション
    Posts []Post `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

// TableName テーブル名を指定
func (Tag) TableName() string {
    return "tags"
}

// TagForCreate タグ作成用構造体
type TagForCreate struct {
    Name string `json:"name" binding:"required,min=1,max=100"`
}