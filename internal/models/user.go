// internal/models/user.go (拡張版)
package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User ユーザーモデル
type User struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string     `gorm:"size:255;not null;index:idx_user_name" json:"name" validate:"required,min=1,max=255"`
	Email           string     `gorm:"size:255;uniqueIndex:idx_user_email;not null" json:"email" validate:"required,email,max=255"`
	EmailVerifiedAt *time.Time `gorm:"null" json:"email_verified_at"`
	Password        string     `gorm:"size:255;not null" json:"-" validate:"required,min=6"`
	RememberToken   *string    `gorm:"size:100;null;index:idx_user_remember_token" json:"-"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;index:idx_user_created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// リレーション（パフォーマンス最適化）
	Posts    []Post    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"posts,omitempty" validate:"-"`
	Comments []Comment `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"comments,omitempty" validate:"-"`
}

// TableName テーブル名を明示的に指定
func (User) TableName() string {
	return "users"
}

// UserForCreate ユーザー作成用構造体
type UserForCreate struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserForUpdate ユーザー更新用構造体
type UserForUpdate struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Email *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
}

// UserResponse API レスポンス用構造体
type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	PostCount int       `json:"post_count,omitempty"`
}

// UserStats ユーザー統計情報
type UserStats struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PostCount    int    `json:"post_count"`
	CommentCount int    `json:"comment_count"`
}

// BeforeCreate 作成前処理（パスワードハッシュ化）
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword パスワード確認
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Validate バリデーション実行
func (u *User) Validate() error {
	return ValidateStruct(u)
}

// IsEmailVerified メール認証済みか確認
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
