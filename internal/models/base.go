package models

import (
    "time"
	"github.com/go-playground/validator/v10"
)

// BaseModel 全モデル共通のフィールド
type BaseModel struct {
    ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// SoftDeleteModel ソフトデリート対応モデル
type SoftDeleteModel struct {
    BaseModel
    DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Validator グローバルバリデーター
var Validator *validator.Validate

func init() {
    Validator = validator.New()
}

// ValidateStruct 構造体バリデーション
func ValidateStruct(s interface{}) error {
    return Validator.Struct(s)
}