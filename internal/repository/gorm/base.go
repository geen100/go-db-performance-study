// internal/repository/gorm/base.go
package gorm_repo

import (
    "gorm.io/gorm"
)

// BaseRepository 基底リポジトリ
type BaseRepository struct {
    db *gorm.DB
}

// NewBaseRepository 基底リポジトリを作成
func NewBaseRepository(db *gorm.DB) *BaseRepository {
    return &BaseRepository{db: db}
}

// GetDB データベース接続を取得
func (r *BaseRepository) GetDB() *gorm.DB {
    return r.db
}

// WithTransaction トランザクション実行
func (r *BaseRepository) WithTransaction(fn func(*gorm.DB) error) error {
    return r.db.Transaction(fn)
}