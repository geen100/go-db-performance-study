// internal/models/comment.go
package models

import "time"

// Comment コメントモデル
type Comment struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    PostID    uint      `gorm:"not null;index" json:"post_id"`
    UserID    uint      `gorm:"not null;index" json:"user_id"`
    Body      string    `gorm:"type:text;not null" json:"body"`
    CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // リレーション
    Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
    User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName テーブル名を指定
func (Comment) TableName() string {
    return "comments"
}

// CommentForCreate コメント作成用構造体
type CommentForCreate struct {
    PostID uint   `json:"post_id" binding:"required"`
    UserID uint   `json:"user_id" binding:"required"`
    Body   string `json:"body" binding:"required,min=1"`
}

// CommentResponse API レスポンス用構造体
type CommentResponse struct {
    ID        uint         `json:"id"`
    Body      string       `json:"body"`
    CreatedAt time.Time    `json:"created_at"`
    User      UserResponse `json:"user"`
}